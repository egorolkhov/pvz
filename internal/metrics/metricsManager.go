package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)


//go:generate mockgen -source=metricsManager.go -destination=mocks/metrics_mock.go -package=mocks

type Metrics interface {
	IncHTTPRequests(method, endpoint, status string)
	ObserveHTTPRequestDuration(method, endpoint string, duration time.Duration)

	IncPVZCreated()
	IncReceptionsCreated()
	IncProductsCreated()
}

const (
	PvzCreated        = "pvz_created_total"
	ReceptionsCreated = "receptions_created_total"
	ProductsCreated   = "products_created_total"

	HttpRequestTotal    = "http_requests_total"
	HttpDurationSeconds = "http_request_duration_seconds"
)

type MetricsManager struct {
	counters    map[string]prometheus.Counter
	counterVecs map[string]*prometheus.CounterVec
	histograms  map[string]*prometheus.HistogramVec
}

func NewMetricsManager() *MetricsManager {
	return &MetricsManager{
		counters:    make(map[string]prometheus.Counter),
		counterVecs: make(map[string]*prometheus.CounterVec),
		histograms:  make(map[string]*prometheus.HistogramVec),
	}
}

func (m *MetricsManager) RegisterCounter(name, help string) {
	m.counters[name] = promauto.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
}

func (m *MetricsManager) RegisterCounterHttp(name, help string, labels []string) {
	m.counterVecs[name] = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labels)
}

func (m *MetricsManager) RegisterHistogram(name, help string, labels []string, buckets []float64) {
	m.histograms[name] = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labels)
}

func (m *MetricsManager) IncHTTPRequests(method, endpoint, status string) {
	if cv, ok := m.counterVecs[HttpRequestTotal]; ok {
		cv.With(prometheus.Labels{
			"method":   method,
			"endpoint": endpoint,
			"status":   status,
		}).Inc()
	}
}

func (m *MetricsManager) ObserveHTTPRequestDuration(method, endpoint string, duration time.Duration) {
	if hist, ok := m.histograms[HttpDurationSeconds]; ok {
		hist.With(prometheus.Labels{
			"method":   method,
			"endpoint": endpoint,
		}).Observe(duration.Seconds())
	}
}

func (m *MetricsManager) IncPVZCreated() {
	if counter, ok := m.counters[PvzCreated]; ok {
		counter.Inc()
	}
}

func (m *MetricsManager) IncReceptionsCreated() {
	if counter, ok := m.counters[ReceptionsCreated]; ok {
		counter.Inc()
	}
}

func (m *MetricsManager) IncProductsCreated() {
	if counter, ok := m.counters[ProductsCreated]; ok {
		counter.Inc()
	}
}

func Init() *MetricsManager {
	mm := NewMetricsManager()

	mm.RegisterCounterHttp("http_requests_total", "Total number of HTTP requests", []string{"method", "endpoint", "status"})
	mm.RegisterHistogram("http_request_duration_seconds", "Duration of HTTP requests", []string{"method", "endpoint"}, prometheus.DefBuckets)

	mm.RegisterCounter("pvz_created_total", "Total PVZ number")
	mm.RegisterCounter("receptions_created_total", "Total receptions number")
	mm.RegisterCounter("products_created_total", "Total products number")

	return mm
}
