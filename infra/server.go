// Package infra contains lowlevel infrastructure details
//
// Date: 2020-11-24
package infra

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	// nolint:gosec
	_ "net/http/pprof" // for remote profiling

	"git.countmax.ru/countmax/wda.back/internal/permissions"
	"git.countmax.ru/countmax/wda.back/internal/permissions/keto"
	"git.countmax.ru/countmax/wda.back/internal/session"
	"git.countmax.ru/countmax/wda.back/internal/session/inmemory"
	"git.countmax.ru/countmax/wda.back/internal/session/kratos"

	"git.countmax.ru/countmax/wda.back/domain"
	"git.countmax.ru/countmax/wda.back/repos"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	requestIDName                   = "request_id"
	periodHealthCheck time.Duration = 30 * time.Second
	apiVersion        string        = "v2"
	apiLocalVersion   string        = "v1"
	scope             string        = "countmax523"
	scopeUPStream     string        = "layoutconfig.api"
	maxIdleConns      int           = 50
	kindManagerInMem  string        = "memory"
)

// Server main engine
type Server struct {
	log         *zap.SugaredLogger
	consul      *consulapi.Agent
	mux         *echo.Echo
	metricmux   *echo.Echo
	version     string
	githash     string
	build       string
	config      *viper.Viper
	mService    *prometheus.GaugeVec
	mAPI        *prometheus.HistogramVec
	chCancel    <-chan struct{}
	fnCancel    context.CancelFunc
	sess        session.ManagerInterface
	perm        permissions.ManagerInterface
	repo        domain.UserRepoI
	handler     *http.Client
	guiSettings Settings
}

// NewServer builder main document server
func NewServer(version, build, githash string) *Server {
	s := &Server{
		version:  version,
		githash:  githash,
		build:    build,
		mService: common,
		mAPI:     httpDuration,
	}
	// fill config
	s.setConfig()
	s.setLogger(version, build, githash)

	// set session manager
	err := s.setSessManager()
	if err != nil {
		s.log.Fatalf("failed %s", err)
	}
	err = s.setPermManager()
	if err != nil {
		s.log.Fatalf("failed %s", err)
	}
	//
	s.mService.WithLabelValues("general", "localhost", s.version, s.githash, s.build).Set(0)
	s.registerRepos()
	// start healthChecker
	go s.healthChecker(periodHealthCheck, s.chCancel)

	return s
}

// Run is running Server
func (s *Server) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	s.fnCancel = cancel
	s.chCancel = ctx.Done()
	s.registerRoutes()
	s.mux.HTTPErrorHandler = s.customHTTPErrorHandler

	host := s.config.GetString("httpd.host") + ":" + s.config.GetString("httpd.port")
	s.log.Infof("http server starting main service on the [%s] tcp port", host)
	go func() {
		if err := s.mux.Start(host); !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatalf("http server error: %v", err)
		}
	}()
	// metrics
	hostportMetrics := s.config.GetString("httpd.host") + ":" + s.config.GetString("httpd.service.port")
	s.log.Infof("http server starting service api on the [%s]", hostportMetrics)
	go func() {
		if err := s.metricmux.Start(hostportMetrics); !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatalf("http server error: %v", err)
		}
	}()
	s.consulRegister()
}

// Stop is stopping Server
func (s *Server) Stop() {
	s.log.Infof("got signal to stopping server")
	stopDuration := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), stopDuration)
	defer cancel()
	defer s.consulDeRegister()
	s.fnCancel()
	go func(ctx context.Context) {
		if err := s.mux.Shutdown(ctx); err != nil {
			s.log.Fatal(err)
		}
	}(ctx)
	if err := s.metricmux.Shutdown(ctx); err != nil {
		s.log.Fatal(err)
	}
}

func (s *Server) registerRoutes() {
	// service metrics
	me := echo.New()
	me.HidePort = true
	me.HideBanner = true // hide banner ECHO
	me.Use(middleware.Recover())
	me.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	me.GET("/health", s.apiHealthCheck)
	//
	// pprof
	dbg := me.Group("/debug")
	dbg.GET("/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	s.metricmux = me
	// main
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true // hide banner ECHO
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(s.customHTTPLogger)

	uri := s.config.GetString("proxy.upstream")
	url1, err := url.Parse(uri)
	if err != nil {
		s.log.Fatalf("parse url %s error, %v", uri, err)
	}
	proxyTimeout := s.config.GetDuration("proxy.timeout_sec") * time.Second
	// nolint:gosec
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	// set timeouts and border for opened connections
	handler := &http.Client{Timeout: proxyTimeout, Transport: &http.Transport{
		MaxIdleConns:    maxIdleConns,
		IdleConnTimeout: proxyTimeout,
		TLSClientConfig: cfg,
	}}
	s.handler = handler
	// common
	balancer := middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: url1,
		},
	})

	v2 := e.Group("/" + apiVersion)
	v2.Use(s.checkSession)
	v2.Use(middleware.Proxy(balancer))

	// v1
	v1 := e.Group("/" + apiLocalVersion)
	// settings
	v1.GET("/layout/settings", s.apiSettings)
	// static
	e.Static("/", "web")
	s.mux = e
}

func (s *Server) registerRepos() {
	s.log.Debug("start registerRepos")
	defer s.log.Debug("finish registerRepos")
	timeout := s.config.GetDuration("countmax.timeout_sec") * time.Second
	dsn := s.config.GetString("countmax.url")

	cmr, err := repos.NewCMRepo(dsn, timeout, s.log)
	if err != nil {
		s.log.Fatalf("registerRepo by config error, %v", err)
	}
	s.repo = cmr
	s.mService.WithLabelValues(scope, cmr.GetSrvPortDB(), s.version, s.githash, s.build).Set(1)

	// settings
	proxy := s.config.GetString("layout.proxy")
	vOnline := s.config.GetString("layout.visible.online")
	vQueue := s.config.GetString("layout.visible.queue")
	vReport := s.config.GetString("layout.visible.report")
	vMonitor := s.config.GetString("layout.visible.monitoring")
	s.guiSettings = Settings{
		Proxy:             proxy,
		VisibleOnline:     vOnline,
		VisibleQueue:      vQueue,
		VisibleReport:     vReport,
		VisibleMonitoring: vMonitor,
	}
}

func (s *Server) setSessManager() error {
	sessKind := s.config.GetString("session.source")
	if sessKind == "" || sessKind != "kratos" {
		sessKind = kindManagerInMem
	}
	switch sessKind {
	case kindManagerInMem:
		ms := inmemory.New()
		s.sess = ms
		return nil
	case "kratos":
		kratosURL := s.config.GetString("session.url")
		kratosTIO := s.config.GetDuration("session.timeout")
		km, err := kratos.New(kratosURL, kratosTIO, s.log, httpDuration)
		if err != nil {
			return err
		}
		s.sess = km
		return nil
	default:
		return errors.New("doesn't defined kind of the session manager")
	}
}

func (s *Server) setPermManager() error {
	sessKind := s.config.GetString("permissions.source")
	if sessKind == "" || sessKind != "keto" {
		sessKind = kindManagerInMem
	}
	switch sessKind {
	case "keto":
		ketoURL := s.config.GetString("permissions.url")
		ketoTIO := s.config.GetDuration("permissions.timeout")
		km, err := keto.New(ketoURL, ketoTIO, s.log, httpDuration)
		if err != nil {
			return err
		}
		s.perm = km
		return nil
	default:
		return errors.New("doesn't defined kind of the permission manager")
	}
}

func (s *Server) healthChecker(period time.Duration, cancel <-chan struct{}) {
	s.log.Debugf("starting healthChecker")
	defer s.log.Debugf("stopped healthChecker")
	tick := time.NewTicker(period)
	for {
		select {
		case <-cancel:
			tick.Stop()
			return
		case <-tick.C:
			s.log.Debug("time to healthCheck")
			err := s.healthCheck()
			if err != nil {
				s.log.Errorf("healthCheck failed %s", err)
			}
		}
	}
}

func (s *Server) healthCheck() error {
	// service_up general =1 if one of layoutRepo health OK
	s.log.Debugf("starting healthCheck")
	defer s.log.Debugf("stopped healthCheck")
	dest := s.repo.GetSrvPortDB()
	upDest := s.config.GetString("proxy.upstream")
	err := s.repo.HealthCheck()
	if err != nil {
		s.log.Errorf("HealthCheck error, %v", err)
		s.mService.WithLabelValues(scope, dest, s.version, s.githash, s.build).Set(0)
		s.mService.WithLabelValues("general", "localhost", s.version, s.githash, s.build).Set(0)
		return err
	}
	s.mService.WithLabelValues(scope, dest, s.version, s.githash, s.build).Set(1)
	err = s.upstreamHealthCheck()
	if err != nil {
		s.log.Errorf("upstreamHealthCheck error, %v", err)
		s.mService.WithLabelValues(scopeUPStream, upDest, s.version, s.githash, s.build).Set(0)
		s.mService.WithLabelValues("general", "localhost", s.version, s.githash, s.build).Set(0)
		return err
	}
	s.mService.WithLabelValues(scopeUPStream, upDest, s.version, s.githash, s.build).Set(1)
	s.mService.WithLabelValues("general", "localhost", s.version, s.githash, s.build).Set(1)
	return nil
}

func (s *Server) upstreamHealthCheck() error {
	upDest := s.config.GetString("proxy.health")
	uri := fmt.Sprintf("%s/health", upDest)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}
	resp, err := s.handler.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			s.log.Errorf("response body close error %s", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		return fmt.Errorf("uri: %s, bad response, httpcode=%d, body=%v", uri, resp.StatusCode, string(body))
	}
	return nil
}
