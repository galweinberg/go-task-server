// internal/metrics/metrics.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests grouped by path",
		},
		[]string{"path"},
	)
)

// Register registers all Prometheus metrics
func Register() {
	prometheus.MustRegister(httpRequests)
}

// Inc increments request counter for a specific path
func Inc(path string) {
	httpRequests.WithLabelValues(path).Inc()
}
