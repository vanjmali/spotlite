package telemetry

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/trace"
)

// TraceID returns the current trace ID if present.
func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	sc := span.SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

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

// AttachMuxTracing wires OpenTelemetry middleware and the trace header response helper.
// It uses `otelmux` under the hood for the Gorilla Mux library.
func AttachMuxTracing(r *mux.Router, serviceName string) {
	r.Use(otelmux.Middleware(serviceName))
	r.Use(TraceHeaderMiddleware)
	r.Use(logging.HTTPMiddleware)
}
