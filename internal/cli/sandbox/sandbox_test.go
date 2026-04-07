package sandbox

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/bootstrap"
)

func TestTruncateSessionKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{give: "", want: "--------"},
		{give: "abc", want: "abc     "},
		{give: "abcdefgh", want: "abcdefgh"},
		{give: "abcdefghij", want: "abcdefgh"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := truncateSessionKey(tt.give, 8)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, 8, len(got),
				"truncateSessionKey must always return a fixed-width string for column alignment")
		})
	}
}

// TestRenderRecentDecisions_NilBootLoaderSilent verifies that the diagnostic
// remains a pure sandbox-layer inspection tool when no bootstrap is wired:
// passing a nil BootLoader must produce no output (no panic, no header).
func TestRenderRecentDecisions_NilBootLoaderSilent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	renderRecentDecisions(context.Background(), &buf, nil, "")
	assert.Empty(t, buf.String(),
		"nil BootLoader must produce no output so `sandbox status` works without bootstrap")
}

// TestRenderRecentDecisions_BootLoaderErrorSilent verifies that a bootstrap
// failure (DB locked, signed-out, missing) is swallowed silently — Recent
// Decisions is best-effort and must not break the rest of the status command.
func TestRenderRecentDecisions_BootLoaderErrorSilent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	loader := func() (*bootstrap.Result, error) {
		return nil, errors.New("database locked")
	}
	renderRecentDecisions(context.Background(), &buf, loader, "")
	assert.Empty(t, buf.String(),
		"BootLoader error must be swallowed silently — graceful degradation")
}

// TestRenderRecentDecisions_NilDBClientSilent verifies that a bootstrap
// result with no DB client (e.g. opened in degraded mode) is also silent.
func TestRenderRecentDecisions_NilDBClientSilent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	loader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	}
	renderRecentDecisions(context.Background(), &buf, loader, "")
	assert.Empty(t, buf.String(),
		"nil DBClient must be swallowed silently")
}
