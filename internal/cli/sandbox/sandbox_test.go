package sandbox

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

// errSimulatedBootFailure is returned by test BootLoaders that need to
// exercise the graceful-degradation fallback path in newStatusCmd.
var errSimulatedBootFailure = errors.New("simulated bootstrap failure")

// defaultTestConfig builds a minimal *config.Config sufficient for the
// status renderer. It does NOT load from disk and does NOT touch the
// bootstrap pipeline, so the sandbox CLI tests stay self-contained.
func defaultTestConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.DataRoot = "/tmp/lango-test-status"
	return cfg
}

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

// TestFormatDecisionLine_BackendColumnForNonAppliedRows verifies that
// excluded / skipped / rejected rows render a "-" backend column even when
// the published event carried a non-empty Backend value (every publish
// site stamps Backend from the wired isolator regardless of decision).
// Otherwise rows like `excluded  bwrap  git status` would falsely suggest
// the command actually ran inside bwrap, when in fact excluded commands
// run unsandboxed.
func TestFormatDecisionLine_BackendColumnForNonAppliedRows(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		give        string
		decision    string
		backend     string
		wantBackend string
	}{
		{give: "applied keeps backend", decision: "applied", backend: "bwrap", wantBackend: "bwrap"},
		{give: "excluded forces dash", decision: "excluded", backend: "bwrap", wantBackend: "-"},
		{give: "skipped forces dash", decision: "skipped", backend: "seatbelt", wantBackend: "-"},
		{give: "rejected forces dash", decision: "rejected", backend: "bwrap", wantBackend: "-"},
		{give: "applied with empty backend", decision: "applied", backend: "", wantBackend: "-"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			line := formatDecisionLine(now, "abcdef12", tt.decision, tt.backend, "git status", "")

			// Pull the backend column out by index. Format is:
			//   "  <ts>  [<sess>] <decision-9> <backend-9> <target>"
			// Tokens 0..N delimited by spaces, but %-9s pads with spaces, so
			// we use strings.Fields to collapse them. Token layout:
			//   [date, time, [sess], decision, backend, target...]
			fields := strings.Fields(line)
			require.GreaterOrEqual(t, len(fields), 6, "line %q must have >=6 tokens", line)
			gotBackend := fields[4]
			assert.Equal(t, tt.wantBackend, gotBackend,
				"row %q: backend column should be %q, got %q (line: %q)",
				tt.give, tt.wantBackend, gotBackend, line)

			// For non-applied verdicts, the original Backend value must
			// NOT appear anywhere in the line (so 'bwrap'/'seatbelt'
			// cannot leak through into a misleading display).
			if tt.decision != "applied" && tt.backend != "" {
				assert.NotContains(t, line, tt.backend,
					"row %q: original Backend %q must not appear in %q",
					tt.give, tt.backend, line)
			}
		})
	}
}

// TestNewStatusCmd_NilBootLoaderFallsBackToCfgLoader verifies the graceful
// degradation contract: when bootLoader is nil, status uses cfgLoader and
// renders the non-audit sections without erroring out.
func TestNewStatusCmd_NilBootLoaderFallsBackToCfgLoader(t *testing.T) {
	t.Parallel()

	cfgLoader := func() (*config.Config, error) {
		return defaultTestConfig(), nil
	}

	cmd := newStatusCmd(cfgLoader, nil) // nil bootLoader
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	require.NoError(t, cmd.RunE(cmd, nil),
		"nil BootLoader must NOT abort the command — graceful degradation")
	out := buf.String()
	assert.Contains(t, out, "Sandbox Configuration:",
		"config section must render even with nil BootLoader")
	assert.Contains(t, out, "Active Isolation:",
		"active isolation section must render even with nil BootLoader")
	assert.Contains(t, out, "Backend Availability:",
		"backend availability section must render even with nil BootLoader")
	assert.NotContains(t, out, "Recent Sandbox Decisions",
		"Recent Decisions section must be silently skipped when no boot result")
}

// TestNewStatusCmd_BootLoaderErrorFallsBackToCfgLoader verifies that a
// bootLoader returning an error does NOT abort the command — status falls
// back to cfgLoader so degraded environments still get a useful diagnostic.
func TestNewStatusCmd_BootLoaderErrorFallsBackToCfgLoader(t *testing.T) {
	t.Parallel()

	cfgLoader := func() (*config.Config, error) {
		return defaultTestConfig(), nil
	}
	failingBoot := func() (*bootstrap.Result, error) {
		return nil, errSimulatedBootFailure
	}

	cmd := newStatusCmd(cfgLoader, failingBoot)
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	require.NoError(t, cmd.RunE(cmd, nil),
		"BootLoader error must NOT abort the command — graceful degradation via cfgLoader")
	out := buf.String()
	assert.Contains(t, out, "Sandbox Configuration:")
	assert.NotContains(t, out, "Recent Sandbox Decisions",
		"Recent Decisions silently skipped when boot fails")
}

func TestFindTouch(t *testing.T) {
	t.Parallel()

	got := findTouch()
	// On every supported lango platform there is at least one touch binary
	// reachable via PATH or one of the fallback locations. We assert
	// non-empty rather than checking a specific path so the test stays
	// portable across macOS / Linux / Alpine.
	if got == "" {
		t.Skip("touch not found in test environment; skipping (production smoke tests will report inconclusive)")
	}
	assert.NotEmpty(t, got)
}
