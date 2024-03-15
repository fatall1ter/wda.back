package infra

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setLogger fills log field in server structure
func (s *Server) setLogger(version, build, githash string) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = s.config.GetString("httpd.host")
	}
	// set log output
	logOutput := os.Stdout
	if s.config.GetString("log.file") != "" && s.config.GetString("log.file") != "stdout" {
		f, err := os.Create(s.config.GetString("log.file"))
		if err != nil {
			fmt.Printf("open log file [%s] error, %v", s.config.GetString("log.file"), err)
		} else {
			logOutput = f
		}
	}

	atom := zap.NewAtomicLevel()
	switch s.config.GetString("log.level") {
	case "debug", "debugging", "deb", "debag":
		atom.SetLevel(zap.DebugLevel)
	case "info", "information", "inf":
		atom.SetLevel(zap.InfoLevel)
	case "warn", "warning", "WARN":
		atom.SetLevel(zap.WarnLevel)
	case "err", "error":
		atom.SetLevel(zap.ErrorLevel)
	default:
		atom.SetLevel(zap.InfoLevel)
	}

	// To keep the example deterministic, disable timestamps in the output.
	var encoderCfg zapcore.EncoderConfig
	if s.config.GetString("env") == "production" {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}

	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(logOutput),
		atom,
	),
		zap.AddCaller()).With(
		zap.String("program", s.config.GetString("app.name")),
		zap.String("hostname", hostname),
		zap.String("version", version),
		zap.String("build", build),
		zap.String("githash", githash),
		zap.String("env", s.config.GetString("env")))

	s.log = logger.Sugar()
}
