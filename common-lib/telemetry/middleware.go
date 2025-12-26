package telemetry

import "net/http"

// TraceHeaderMiddleware adds the current trace ID to responses for easier debugging.
func TraceHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := TraceID(r.Context())
		if traceID != "" {
			w.Header().Set("X-Trace-Id", traceID)
		}
		next.ServeHTTP(w, r)
	})
}
