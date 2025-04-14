package middleware

import (
	"avito/internal/metrics"
	"net/http"
	"strconv"
	"time"
)

// statusRecorder поможет «схватить» статус код ответа
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(m metrics.Metrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			method := r.Method
			path := r.URL.Path

			status := strconv.Itoa(rec.statusCode)

			m.IncHTTPRequests(method, path, status)
			m.ObserveHTTPRequestDuration(method, path, duration)
		})
	}
}
