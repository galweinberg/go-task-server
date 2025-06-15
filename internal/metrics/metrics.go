// internal/metrics/metrics.go
package metrics

import (
	"log"
	"github.com/prometheus/client_golang/prometheus"
)
var (
	taskSubmitted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "task_submitted_total",
			Help: "Total number of tasks submitted by clients",
		},
	)
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
	prometheus.MustRegister(taskSubmitted)

	// Ensure the counter is initialized so Prometheus sees it
	taskSubmitted.Add(0)
}


func IncSubmitted() {
	taskSubmitted.Inc()
log.Println("🔥 IncSubmitted called")
}

// Inc increments request counter for a specific path
func Inc(path string) {
	httpRequests.WithLabelValues(path).Inc()
}
