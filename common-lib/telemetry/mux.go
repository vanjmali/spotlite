package telemetry

import (
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

// AttachMuxTracing wires OpenTelemetry middleware and the trace header response helper.
// It uses `otelmux` under the hood for the Gorilla Mux library.
func AttachMuxTracing(r *mux.Router, serviceName string) {
	r.Use(otelmux.Middleware(serviceName))
	r.Use(TraceHeaderMiddleware)
}
