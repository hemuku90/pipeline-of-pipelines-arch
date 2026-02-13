package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// Business metrics
	usersTotal       prometheus.Counter
	operationsTotal  *prometheus.CounterVec
	buildsTotal      prometheus.Counter
	deploymentsTotal *prometheus.CounterVec

	// Runtime metrics
	goroutines prometheus.Gauge
	memory     *prometheus.GaugeVec

	// Server
	serverName string
	metricsPort int
}

// New creates a new Metrics instance
func New(name string, port int) *Metrics {
	m := &Metrics{
		serverName:  name,
		metricsPort: port,
	}

	// HTTP metrics
	m.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"method", "path", "status"},
	)

	m.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request duration in seconds",
			ConstLabels: prometheus.Labels{"service": name},
			Buckets:     []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	m.httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name:        "http_requests_in_flight",
			Help:        "Number of HTTP requests currently being processed",
			ConstLabels: prometheus.Labels{"service": name},
		},
	)

	// Business metrics
	m.usersTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name:        "app_users_total",
			Help:        "Total number of users created",
			ConstLabels: prometheus.Labels{"service": name},
		},
	)

	m.operationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "app_operations_total",
			Help:        "Total number of operations by type",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"operation", "status"},
	)

	m.buildsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name:        "app_builds_total",
			Help:        "Total number of builds",
			ConstLabels: prometheus.Labels{"service": name},
		},
	)

	m.deploymentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "app_deployments_total",
			Help:        "Total number of deployments by environment",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"environment", "status"},
	)

	// Runtime metrics
	m.goroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name:        "app_goroutines",
			Help:        "Number of active goroutines",
			ConstLabels: prometheus.Labels{"service": name},
		},
	)

	m.memory = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			GoCollectInterval: 10 * time.Second,
			Name:              "app_memory_bytes",
			Help:              "Memory usage in bytes",
			ConstLabels:       prometheus.Labels{"service": name},
		},
		[]string{"type"},
	)

	return m
}

// IncRequest increments the request counter
func (m *Metrics) IncRequest(path string) {
	m.httpRequestsTotal.WithLabelValues("unknown", path, "200").Inc()
}

// IncRequestWithStatus increments the request counter with status
func (m *Metrics) IncRequestWithStatus(method, path, status string) {
	m.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
}

// ObserveDuration records the request duration
func (m *Metrics) ObserveDuration(method, path string, duration time.Duration) {
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// IncUsers increments the user counter
func (m *Metrics) IncUsers() {
	m.usersTotal.Inc()
}

// IncOperation increments the operation counter
func (m *Metrics) IncOperation(operation, status string) {
	m.operationsTotal.WithLabelValues(operation, status).Inc()
}

// IncBuilds increments the build counter
func (m *Metrics) IncBuilds() {
	m.buildsTotal.Inc()
}

// IncDeployments increments the deployment counter
func (m *Metrics) IncDeployments(environment, status string) {
	m.deploymentsTotal.WithLabelValues(environment, status).Inc()
}

// UpdateGoroutines updates the goroutine count
func (m *Metrics) UpdateGoroutines(count int) {
	m.goroutines.Set(float64(count))
}

// UpdateMemory updates memory metrics
func (m *Metrics) UpdateMemory(heap, stack, gc uint64) {
	m.memory.WithLabelValues("heap").Set(float64(heap))
	m.memory.WithLabelValues("stack").Set(float64(stack))
	m.memory.WithLabelValues("gc").Set(float64(gc))
}

// Handler returns the Prometheus metrics HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// MetricsServer returns the metrics HTTP server
func (m *Metrics) MetricsServer() *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return &http.Server{
		Addr:    ":" + strconv.Itoa(m.metricsPort),
		Handler: mux,
	}
}

// Middleware returns a middleware for collecting HTTP metrics
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.httpRequestsInFlight.Inc()
		defer m.httpRequestsInFlight.Dec()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		status := strconv.Itoa(200) // In production, get actual status
		m.IncRequestWithStatus(r.Method, r.URL.Path, status)
		m.ObserveDuration(r.Method, r.URL.Path, duration)
	})
}

// RecordHTTPOutcome records the outcome of an HTTP request
func (m *Metrics) RecordHTTPOutcome(method, path, statusCode string, duration time.Duration) {
	m.httpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}