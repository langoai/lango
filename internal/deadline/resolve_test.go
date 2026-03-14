package deadline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveTimeouts_DefaultFixedTimeout(t *testing.T) {
	t.Parallel()

	idle, ceiling := ResolveTimeouts(TimeoutConfig{
		RequestTimeout: 5 * time.Minute,
	})
	assert.Equal(t, time.Duration(0), idle, "idle should be disabled by default")
	assert.Equal(t, 5*time.Minute, ceiling, "ceiling should equal RequestTimeout")
}

func TestResolveTimeouts_DefaultNoConfig(t *testing.T) {
	t.Parallel()

	idle, ceiling := ResolveTimeouts(TimeoutConfig{})
	assert.Equal(t, time.Duration(0), idle, "idle should be disabled")
	assert.Equal(t, 5*time.Minute, ceiling, "ceiling should default to 5m")
}

func TestResolveTimeouts_ExplicitIdleTimeout(t *testing.T) {
	t.Parallel()

	idle, ceiling := ResolveTimeouts(TimeoutConfig{
		IdleTimeout:    2 * time.Minute,
		RequestTimeout: 30 * time.Minute,
	})
	assert.Equal(t, 2*time.Minute, idle, "idle should match IdleTimeout")
	assert.Equal(t, 30*time.Minute, ceiling, "ceiling should match RequestTimeout")
}

func TestResolveTimeouts_IdleTimeoutWithoutRequestTimeout(t *testing.T) {
	t.Parallel()

	idle, ceiling := ResolveTimeouts(TimeoutConfig{
		IdleTimeout: 2 * time.Minute,
	})
	assert.Equal(t, 2*time.Minute, idle, "idle should match IdleTimeout")
	assert.Equal(t, 60*time.Minute, ceiling, "ceiling should default to 60m when idle is set")
}

func TestResolveTimeouts_IdleTimeoutExplicitlyDisabled(t *testing.T) {
	t.Parallel()

	idle, ceiling := ResolveTimeouts(TimeoutConfig{
		IdleTimeout:    -1,
		RequestTimeout: 10 * time.Minute,
	})
	assert.Equal(t, time.Duration(0), idle, "idle should be disabled when set to -1")
	assert.Equal(t, 10*time.Minute, ceiling, "ceiling should equal RequestTimeout")
}

func TestResolveTimeouts_LegacyAutoExtend(t *testing.T) {
	t.Parallel()

	idle, ceiling := ResolveTimeouts(TimeoutConfig{
		AutoExtendTimeout: true,
		RequestTimeout:    5 * time.Minute,
		MaxRequestTimeout: 15 * time.Minute,
	})
	assert.Equal(t, 5*time.Minute, idle, "idle should equal RequestTimeout in legacy mode")
	assert.Equal(t, 15*time.Minute, ceiling, "ceiling should equal MaxRequestTimeout")
}

func TestResolveTimeouts_LegacyAutoExtendDefaults(t *testing.T) {
	t.Parallel()

	idle, ceiling := ResolveTimeouts(TimeoutConfig{
		AutoExtendTimeout: true,
	})
	assert.Equal(t, 5*time.Minute, idle, "idle should default to 5m")
	assert.Equal(t, 15*time.Minute, ceiling, "ceiling should default to 3x idle")
}

func TestResolveTimeouts_CeilingMinGreaterThanIdle(t *testing.T) {
	t.Parallel()

	// Edge case: RequestTimeout <= IdleTimeout
	idle, ceiling := ResolveTimeouts(TimeoutConfig{
		IdleTimeout:    10 * time.Minute,
		RequestTimeout: 5 * time.Minute,
	})
	assert.Equal(t, 10*time.Minute, idle)
	assert.Greater(t, ceiling, idle, "ceiling must be greater than idle timeout")
}
