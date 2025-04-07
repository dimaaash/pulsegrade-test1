package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// MetricsMiddleware wraps an HTTP handler with metrics instrumentation
func MetricsMiddleware(next http.Handler, environment string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response writer wrapper to capture the status code
		rww := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200 OK
		}

		// Record start time
		startTime := time.Now()

		// Call the next handler
		next.ServeHTTP(rww, r)

		// Record duration
		duration := time.Since(startTime).Seconds()

		// Extract path for the metrics label (could be improved with route pattern matching)
		endpoint := r.URL.Path

		// Record metrics
		HttpRequestsTotal.WithLabelValues(endpoint, r.Method, strconv.Itoa(rww.statusCode), environment).Inc()
		HttpRequestDuration.WithLabelValues(endpoint, r.Method, environment).Observe(duration)
	})
}

// responseWriterWrapper is a custom response writer that captures the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before calling the wrapped ResponseWriter
func (rww *responseWriterWrapper) WriteHeader(statusCode int) {
	rww.statusCode = statusCode
	rww.ResponseWriter.WriteHeader(statusCode)
}
