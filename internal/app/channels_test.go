package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/config"
)

// TestResolveTimeouts_DelegatesToDeadlinePackage verifies that App.resolveTimeouts()
// correctly delegates to deadline.ResolveTimeouts with matching results.
// The exhaustive logic tests live in internal/deadline/resolve_test.go.
func TestResolveTimeouts_DelegatesToDeadlinePackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         config.AgentConfig
		wantIdle    time.Duration
		wantCeiling time.Duration
	}{
		{
			name:        "default fixed timeout",
			cfg:         config.AgentConfig{RequestTimeout: 5 * time.Minute},
			wantIdle:    0,
			wantCeiling: 5 * time.Minute,
		},
		{
			name:        "explicit idle timeout",
			cfg:         config.AgentConfig{IdleTimeout: 2 * time.Minute, RequestTimeout: 30 * time.Minute},
			wantIdle:    2 * time.Minute,
			wantCeiling: 30 * time.Minute,
		},
		{
			name:        "legacy auto-extend",
			cfg:         config.AgentConfig{AutoExtendTimeout: true, RequestTimeout: 5 * time.Minute, MaxRequestTimeout: 15 * time.Minute},
			wantIdle:    5 * time.Minute,
			wantCeiling: 15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := &App{Config: &config.Config{Agent: tt.cfg}}
			idle, ceiling := app.resolveTimeouts()
			assert.Equal(t, tt.wantIdle, idle)
			assert.Equal(t, tt.wantCeiling, ceiling)
		})
	}
}
