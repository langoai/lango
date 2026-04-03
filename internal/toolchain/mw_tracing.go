package toolchain

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/langoai/lango/internal/agent"
)

// WithTracing returns a middleware that wraps each tool invocation in an
// OpenTelemetry span. The span records the tool name, parameter count,
// and any error. It should be placed as the outermost middleware so that
// blocked calls (by policy/approval) are also traced.
func WithTracing(tracer trace.Tracer) Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			ctx, span := tracer.Start(ctx, fmt.Sprintf("tool/%s", tool.Name),
				trace.WithAttributes(
					attribute.String("tool.name", tool.Name),
					attribute.Int("tool.params_count", len(params)),
				),
			)
			defer span.End()

			result, err := next(ctx, params)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}

			return result, err
		}
	}
}
