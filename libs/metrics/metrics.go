package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

func NewServiceMetricsWithRegistry(serviceName string, reg prometheus.Registerer) *ServiceMetrics {
	prefix := serviceName + "_"

	requestTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prefix + "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prefix + "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	requestInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: prefix + "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	tasksCreated := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: prefix + "tasks_created_total",
			Help: "Total number of tasks created",
		},
	)

	tasksProcessed := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: prefix + "tasks_processed_total",
			Help: "Total number of tasks processed",
		},
	)

	tasksFailed := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: prefix + "tasks_failed_total",
			Help: "Total number of tasks failed",
		},
	)

	tasksInQueue := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: prefix + "tasks_in_queue",
			Help: "Current number of tasks in queue",
		},
	)

	if reg != nil {
		reg.MustRegister(
			requestTotal,
			requestDuration,
			requestInFlight,
			tasksCreated,
			tasksProcessed,
			tasksFailed,
			tasksInQueue,
		)
	}

	return &ServiceMetrics{
		RequestTotal:    requestTotal,
		RequestDuration: requestDuration,
		RequestInFlight: requestInFlight,
		TasksCreated:    tasksCreated,
		TasksProcessed:  tasksProcessed,
		TasksFailed:     tasksFailed,
		TasksInQueue:    tasksInQueue,
	}
}

func NewServiceMetrics(serviceName string) *ServiceMetrics {
	return NewServiceMetricsWithRegistry(serviceName, prometheus.DefaultRegisterer)
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
