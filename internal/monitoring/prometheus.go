package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// RequestsTotal tracks the total number of HTTP requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method and path",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration tracks the duration of HTTP requests
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// DatabaseConnections tracks the current number of active DB connections
	DatabaseConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Number of active database connections",
		},
	)

	// ClicksLogged tracks the rate of clicks being logged
	ClicksLogged = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "clicks_logged_total",
			Help: "Total number of ad clicks logged",
		},
	)
)

// MetricsHandler returns a handler for the /metrics endpoint
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// PrometheusMiddleware adds metrics to HTTP requests
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		wrapped := newResponseWriter(w)

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Record metrics after the request is complete
		duration := time.Since(start).Seconds()
		statusCode := wrapped.statusCode

		RequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(statusCode)).Inc()
		RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// SetDatabaseConnections sets the current number of database connections
func SetDatabaseConnections(count int) {
	DatabaseConnections.Set(float64(count))
}

// IncrementClicksLogged increases the clicks logged counter
func IncrementClicksLogged() {
	ClicksLogged.Inc()
}
