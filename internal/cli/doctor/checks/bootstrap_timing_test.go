package checks

import (
	"context"
	"testing"
	"time"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapTimingCheck_NoBoot(t *testing.T) {
	t.Parallel()
	c := &BootstrapTimingCheck{}
	r := c.RunWithBootstrap(context.Background(), nil, nil)
	assert.Equal(t, StatusSkip, r.Status)
}

func TestBootstrapTimingCheck_NoBaseline(t *testing.T) {
	t.Parallel()

	boot := &bootstrap.Result{
		PhaseTiming: []bootstrap.PhaseTimingEntry{
			{Phase: "config", Duration: 10 * time.Millisecond},
		},
	}

	c := &BootstrapTimingCheck{}
	r := c.RunWithBootstrap(context.Background(), nil, boot)
	// Result depends on whether ~/.lango/diagnostics has existing data.
	// In a clean environment: Skip (no baseline). With existing data: Pass or Warn.
	assert.NotEqual(t, StatusFail, r.Status)
}

func TestBootstrapTimingCheck_Run_FallbackSkip(t *testing.T) {
	t.Parallel()
	c := &BootstrapTimingCheck{}
	r := c.Run(context.Background(), nil)
	assert.Equal(t, StatusSkip, r.Status)
}

func TestComputeMedians(t *testing.T) {
	t.Parallel()

	entries := []bootstrap.TimingLogEntry{
		{Phases: []bootstrap.PhaseTimingRecord{{Name: "a", DurationMs: 10}, {Name: "b", DurationMs: 100}}},
		{Phases: []bootstrap.PhaseTimingRecord{{Name: "a", DurationMs: 20}, {Name: "b", DurationMs: 200}}},
		{Phases: []bootstrap.PhaseTimingRecord{{Name: "a", DurationMs: 30}, {Name: "b", DurationMs: 300}}},
	}

	m := computeMedians(entries)
	assert.Equal(t, int64(20), m["a"])
	assert.Equal(t, int64(200), m["b"])
}
