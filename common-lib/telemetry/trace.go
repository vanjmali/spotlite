package telemetry

import (
	"context"

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
