package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/toolchain"
)

// wireTestObservability creates a minimal EventBus + MetricsCollector wired
// with the same subscriptions as the real app, then returns both so tests
// can publish events and assert on collector state.
func wireTestObservability(t *testing.T) (*eventbus.Bus, *observability.MetricsCollector) {
	t.Helper()

	bus := eventbus.New()
	collector := observability.NewCollector()

	cfg := config.DefaultConfig()
	cfg.Observability.Metrics.Enabled = true

	oc := &observabilityComponents{
		collector: collector,
	}

	// Replicate the wiring from wiring_observability.go (steps 5 & 6).
	eventbus.SubscribeTyped[eventbus.PolicyDecisionEvent](bus, func(evt eventbus.PolicyDecisionEvent) {
		oc.collector.RecordPolicyDecision(evt.Verdict, evt.Reason)
	})

	eventbus.SubscribeTyped[toolchain.ToolExecutedEvent](bus, func(evt toolchain.ToolExecutedEvent) {
		oc.collector.RecordToolExecution(evt.ToolName, evt.AgentName, evt.Duration, evt.Success)
	})

	return bus, collector
}

func TestEventContract_ToolExecuted_IncreasesCount(t *testing.T) {
	t.Parallel()

	bus, collector := wireTestObservability(t)

	snap := collector.Snapshot()
	require.Equal(t, int64(0), snap.ToolExecutions)

	bus.Publish(toolchain.ToolExecutedEvent{
		ToolName:  "exec",
		AgentName: "operator",
		Duration:  100 * time.Millisecond,
		Success:   true,
	})

	snap = collector.Snapshot()
	assert.Equal(t, int64(1), snap.ToolExecutions)
	assert.Contains(t, snap.ToolBreakdown, "exec")
}

func TestEventContract_PolicyDecision_IncreasesBlocks(t *testing.T) {
	t.Parallel()

	bus, collector := wireTestObservability(t)

	snap := collector.Snapshot()
	require.Equal(t, int64(0), snap.Policy.Blocks)

	bus.Publish(eventbus.PolicyDecisionEvent{
		Command: "rm -rf /",
		Verdict: "block",
		Reason:  "catastrophic",
	})

	snap = collector.Snapshot()
	assert.Equal(t, int64(1), snap.Policy.Blocks)
	assert.Equal(t, int64(1), snap.Policy.ByReason["catastrophic"])
}

func TestEventContract_PolicyDecision_IncreasesObserves(t *testing.T) {
	t.Parallel()

	bus, collector := wireTestObservability(t)

	bus.Publish(eventbus.PolicyDecisionEvent{
		Command: "python -c 'print(1)'",
		Verdict: "observe",
		Reason:  "scripting",
	})

	snap := collector.Snapshot()
	assert.Equal(t, int64(0), snap.Policy.Blocks)
	assert.Equal(t, int64(1), snap.Policy.Observes)
}

func TestEventContract_NoEvents_CountersUnchanged(t *testing.T) {
	t.Parallel()

	_, collector := wireTestObservability(t)

	snap := collector.Snapshot()
	assert.Equal(t, int64(0), snap.ToolExecutions)
	assert.Equal(t, int64(0), snap.Policy.Blocks)
	assert.Equal(t, int64(0), snap.Policy.Observes)
	assert.Empty(t, snap.ToolBreakdown)
}

func TestEventContract_MultipleEvents_Accumulate(t *testing.T) {
	t.Parallel()

	bus, collector := wireTestObservability(t)

	for i := 0; i < 5; i++ {
		bus.Publish(toolchain.ToolExecutedEvent{
			ToolName: "fs_read",
			Success:  true,
			Duration: time.Millisecond,
		})
	}
	bus.Publish(toolchain.ToolExecutedEvent{
		ToolName: "fs_read",
		Success:  false,
		Duration: time.Millisecond,
		Error:    "not found",
	})

	snap := collector.Snapshot()
	assert.Equal(t, int64(6), snap.ToolExecutions)

	tm := snap.ToolBreakdown["fs_read"]
	assert.Equal(t, int64(6), tm.Count)
	assert.Equal(t, int64(1), tm.Errors)
}
