package sandbox

import (
	"bytes"
	"context"
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

// TestRenderRecentDecisions_NilBootSilent verifies that the helper is safe
// to call with a nil bootstrap.Result. This matters because the outer
// status command delegates the nil check to the helper for clarity.
func TestRenderRecentDecisions_NilBootSilent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	renderRecentDecisions(context.Background(), &buf, nil, "")
	assert.Empty(t, buf.String(),
		"nil *bootstrap.Result must produce no output")
}

// TestRenderRecentDecisions_NilDBClientSilent verifies that a bootstrap
// result with no DB client (e.g. opened in degraded mode) is also silent.
// Recent Decisions is best-effort and must not break the rest of the status
// command when the audit DB is unavailable.
func TestRenderRecentDecisions_NilDBClientSilent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	renderRecentDecisions(context.Background(), &buf, &bootstrap.Result{}, "")
	assert.Empty(t, buf.String(),
		"nil DBClient must be swallowed silently — graceful degradation")
}
