package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MetricRequestDuration measures the latency of successful requests,
	// grouped by workflow and query.
	MetricRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "drk_request_duration",
		Buckets: []float64{
			0.001, // 1ms
			0.005, // 5ms
			0.01,  // 10ms
			0.025, // 25ms
			0.05,  // 50ms
			0.1,   // 100ms
			0.25,  // 250ms
			0.5,   // 500ms
			1.0,   // 1s
			2.5,   // 2.5s
			5.0,   // 5s
		},
	},
		[]string{
			"workflow",
			"query",
		})

	// MetricErrorDuration measures the latency of failed requests,
	// grouped by workflow and query.
	MetricErrorDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "drk_error_duration",
		Buckets: []float64{
			0.001, // 1ms
			0.005, // 5ms
			0.01,  // 10ms
			0.025, // 25ms
			0.05,  // 50ms
			0.1,   // 100ms
			0.25,  // 250ms
			0.5,   // 500ms
			1.0,   // 1s
			2.5,   // 2.5s
			5.0,   // 5s
		},
	},
		[]string{
			"workflow",
			"query",
		})

	// MetricRequestCount is a running total of the successful requests,
	// grouped by workflow and query.
	MetricRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "drk_request_count",
	},
		[]string{
			"workflow",
			"query",
		})

	// MetricErrorCount is a running total of the failed requests,
	// grouped by workflow and query.
	MetricErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "drk_error_count",
	},
		[]string{
			"workflow",
			"query",
		})
)
