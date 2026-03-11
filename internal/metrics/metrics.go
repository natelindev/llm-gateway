package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metric definitions.
// Metrics are labeled by model name and version for per-model observability.
var (
	// InferTotal counts inference requests by model/version/status.
	InferTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_gateway_infer_total",
			Help: "Total number of inference requests",
		},
		[]string{"model", "version", "status"},
	)

	// InferDuration tracks inference latency distribution in seconds.
	InferDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_gateway_infer_duration_seconds",
			Help:    "Inference request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model", "version"},
	)

	// InferActive reports active inference streams.
	InferActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llm_gateway_infer_active",
			Help: "Current number of active inference streams",
		},
		[]string{"model", "version"},
	)

	// ModelRegistered reports the number of registered model versions.
	ModelRegistered = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llm_gateway_models_registered",
			Help: "Total number of registered model versions",
		},
	)
)
