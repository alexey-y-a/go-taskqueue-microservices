package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ServiceMetrics struct {
	RequestTotal    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	RequestInFlight prometheus.Gauge

	TasksCreated   prometheus.Counter
	TasksProcessed prometheus.Counter
	TasksFailed    prometheus.Counter
	TasksInQueue   prometheus.Gauge
}

func NewServiceMetrics(serviceName string) *ServiceMetrics {
	prefix := serviceName + "_"

	return &ServiceMetrics{
		RequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: prefix + "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),

		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    prefix + "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),

		RequestInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: prefix + "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		TasksCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: prefix + "tasks_created_total",
				Help: "Total number of tasks created",
			},
		),

		TasksProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: prefix + "tasks_processed_total",
				Help: "Total number of tasks processed",
			},
		),

		TasksFailed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: prefix + "tasks_failed_total",
				Help: "Total number of tasks failed",
			},
		),

		TasksInQueue: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: prefix + "tasks_in_queue",
				Help: "Current number of tasks in queue",
			},
		),
	}
}

func (m *ServiceMetrics) MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.RequestInFlight.Inc()
		defer m.RequestInFlight.Dec()

		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(rw, r)

		duration := time.Since(start).Seconds()

		m.RequestTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.statusCode)).Inc()
		m.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
