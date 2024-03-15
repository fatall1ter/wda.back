package infra

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

const (
	configFormat = "yaml"
	fPort        = "p"
	fHost        = "host"
	fLevel       = "level"
	fLogFile     = "logfile"
	fPathConfig  = "c"
	fConsul      = "consul"
	envPrefix    = "WDA"
)

var (
	configName = "config"
	configPath = "."
)

func (s *Server) setConfig() {
	flag.String(fPathConfig, path.Join(configPath, configName+"."+configFormat), "path to config file for application")
	flag.Int(fPort, 8080, "http port for application")
	flag.String(fHost, "0.0.0.0", "ip address for bind application")
	flag.String(fLevel, "debug", "loglevel fog logging information")
	flag.String(fLogFile, "stdout", "log output, can be stdout or file at the disk")
	flag.String(fConsul, "localhost:8500", "consul server address (8500 in original consul server)")
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case fPort:
			_ = os.Setenv(envPrefix+"_HTTPD_PORT", f.Value.String())
		case fHost:
			_ = os.Setenv(envPrefix+"_HTTPD_HOST", f.Value.String())
		case fLevel:
			_ = os.Setenv(envPrefix+"_LOG_LEVEL", f.Value.String())
		case fLogFile:
			_ = os.Setenv(envPrefix+"_LOG_FILE", f.Value.String())
		case fConsul:
			_ = os.Setenv(envPrefix+"_CONSUL_URL", f.Value.String())
		case fPathConfig:
			configPath, configName = path.Split(f.Value.String())
			configName = strings.ReplaceAll(configName, path.Ext(configName), "")
		}
	})

	config := viper.New()
	config.SetConfigType(configFormat)
	config.AddConfigPath(configPath)
	config.SetConfigName(configName)
	config.SetEnvPrefix(envPrefix)
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := config.ReadInConfig() // Find and read the config file
	if err != nil {              // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	config.AutomaticEnv()
	s.config = config
}

// consulRegister register self to cinsul server
func (s *Server) consulRegister() {
	config := consulapi.DefaultConfig()
	config.Address = s.config.GetString("consul.url")
	consul, err := consulapi.NewClient(config)
	if err != nil {
		s.log.Errorf("set consul config for register service error, %v", err)
		return
	}
	s.consul = consul.Agent()
	tags := strings.Split(s.config.GetString("tags"), ",")
	err = consul.Agent().ServiceRegister(
		&consulapi.AgentServiceRegistration{
			ID:                s.config.GetString("consul.serviceid"),
			Name:              s.config.GetString("app.name"),
			Tags:              tags,
			Address:           s.config.GetString("consul.address"),
			Port:              s.config.GetInt("consul.port"),
			Meta:              map[string]string{"version": s.version},
			EnableTagOverride: false,
			Check: &consulapi.AgentServiceCheck{
				DeregisterCriticalServiceAfter: "90m",
				HTTP: fmt.Sprintf("http://%s:%d/health",
					s.config.GetString("consul.address"),
					s.config.GetInt("consul.port")),
				Interval: "60s",
			},
			Weights: &consulapi.AgentWeights{
				Passing: 10,
				Warning: 1,
			},
		})
	if err != nil {
		s.log.Errorf("consul register service error, %v", err)
		return
	}
	s.log.Infof("consul register service success, serviceid=%s, consul.url=%s, consul.address=%s, consul.port=%d",
		s.config.GetString("consul.serviceid"),
		s.config.GetString("consul.url"),
		s.config.GetString("consul.address"),
		s.config.GetInt("consul.port"))
}

// consulDeRegister deregister self from consul server
func (s *Server) consulDeRegister() {
	err := s.consul.ServiceDeregister(s.config.GetString("consul.serviceid"))
	if err != nil {
		s.log.Errorf("can't deregister service %s from consul", s.config.GetString("consul.serviceid"))
		return
	}
	s.log.Infof("success deregister service %s from consul", s.config.GetString("consul.serviceid"))
}
