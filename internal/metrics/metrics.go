package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Technical metrics
	RequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	ResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_time_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	// Business metrics
	PVZCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pvz_created_total",
		Help: "Total number of PVZ created",
	})

	PVZCreateErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pvz_create_errors_total",
		Help: "Total number of PVZ creation errors",
	})

	ReceptionsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "receptions_created_total",
		Help: "Total number of receptions created",
	})

	ReceptionCreateErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reception_create_errors_total",
		Help: "Total number of reception creation errors",
	})

	ProductsAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_added_total",
		Help: "Total number of products added",
	})

	ProductAddErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "product_add_errors_total",
		Help: "Total number of product addition errors",
	})

	// gRPC metrics
	GRPCRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests",
	}, []string{"method", "status"})

	GRPCResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_response_time_seconds",
		Help:    "Duration of gRPC requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}

		defer func() {
			duration := time.Since(start).Seconds()
			ResponseTime.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
			RequestCount.WithLabelValues(r.Method, r.URL.Path, http.StatusText(rw.status)).Inc()
		}()

		next.ServeHTTP(rw, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
