package logging

import (
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// HTTPMiddleware logs request execution with method, path, status and duration.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(sw, r)

		durationMS := time.Since(start).Milliseconds()
		Infof(
			r.Context(),
			"http_request method=%s path=%s status=%d duration_ms=%d remote_addr=%s",
			r.Method,
			r.URL.Path,
			sw.statusCode,
			durationMS,
			r.RemoteAddr,
		)
	})
}
