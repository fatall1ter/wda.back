package infra

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	common = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_up",
			Help: "Common metric of state of available service and subservices",
		},
		[]string{"scope", "destination", "version", "githash", "build"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "api_request_duration_seconds",
			Help: "Summary of api req-resp durations in seconds by quantile",
		},
		[]string{"url", "code", "method"},
	)
)
