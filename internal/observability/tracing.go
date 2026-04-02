package observability

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/langoai/lango/internal/config"
)

// TracerName is the instrumentation name used for Lango spans.
const TracerName = "lango"

// InitTracer initializes an OpenTelemetry TracerProvider based on config.
// Returns the provider and a shutdown function. Caller must call shutdown
// on application exit to flush pending spans.
func InitTracer(cfg config.TracingConfig) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	if !cfg.Enabled {
		// Return a no-op provider.
		tp := sdktrace.NewTracerProvider()
		return tp, tp.Shutdown, nil
	}

	var exporter sdktrace.SpanExporter
	switch cfg.Exporter {
	case "stdout", "":
		exp, err := stdouttrace.New(stdouttrace.WithWriter(os.Stderr))
		if err != nil {
			return nil, nil, fmt.Errorf("create stdout exporter: %w", err)
		}
		exporter = exp
	case "none":
		// No exporter — provider records spans but discards them.
		tp := sdktrace.NewTracerProvider()
		return tp, tp.Shutdown, nil
	default:
		return nil, nil, fmt.Errorf("unsupported tracing exporter: %q", cfg.Exporter)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)

	// Set as global so libraries can pick it up.
	otel.SetTracerProvider(tp)

	return tp, tp.Shutdown, nil
}

// Tracer returns a tracer from the global TracerProvider.
func Tracer() trace.Tracer {
	return otel.Tracer(TracerName)
}
