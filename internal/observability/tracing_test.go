package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestInitTracer_Disabled(t *testing.T) {
	t.Parallel()

	tp, shutdown, err := InitTracer(config.TracingConfig{Enabled: false})
	require.NoError(t, err)
	require.NotNil(t, tp)
	require.NotNil(t, shutdown)

	assert.NoError(t, shutdown(context.Background()))
}

func TestInitTracer_Stdout(t *testing.T) {
	t.Parallel()

	tp, shutdown, err := InitTracer(config.TracingConfig{Enabled: true, Exporter: "stdout"})
	require.NoError(t, err)
	require.NotNil(t, tp)

	// Verify we get a valid tracer.
	tracer := tp.Tracer("test")
	require.NotNil(t, tracer)

	assert.NoError(t, shutdown(context.Background()))
}

func TestInitTracer_None(t *testing.T) {
	t.Parallel()

	tp, shutdown, err := InitTracer(config.TracingConfig{Enabled: true, Exporter: "none"})
	require.NoError(t, err)
	require.NotNil(t, tp)

	assert.NoError(t, shutdown(context.Background()))
}

func TestInitTracer_UnsupportedExporter(t *testing.T) {
	t.Parallel()

	_, _, err := InitTracer(config.TracingConfig{Enabled: true, Exporter: "jaeger"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}
