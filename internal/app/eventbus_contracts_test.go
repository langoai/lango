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

	eventbus.SubscribeTyped[eventbus.GraphAdmissionBatchEvent](bus, func(evt eventbus.GraphAdmissionBatchEvent) {
		oc.collector.RecordGraphAdmissionBatch(observability.GraphAdmissionBatchMetric{
			Source:           evt.Source,
			ProducerGroup:    evt.ProducerGroup,
			ValidatorSource:  evt.ValidatorSource,
			BatchCount:       int64(evt.BatchCount),
			KnownCount:       int64(evt.KnownCount),
			UnknownCount:     int64(evt.UnknownCount),
			UnvalidatedCount: int64(evt.UnvalidatedCount),
		})
	})
	eventbus.SubscribeTyped[eventbus.GraphAdmissionUnmappedSourceEvent](bus, func(evt eventbus.GraphAdmissionUnmappedSourceEvent) {
		oc.collector.RecordGraphAdmissionUnmappedSource(evt.RawSource, int64(evt.BatchCount))
	})
	eventbus.SubscribeTyped[eventbus.GraphExtractorDroppedUnknownEvent](bus, func(evt eventbus.GraphExtractorDroppedUnknownEvent) {
		oc.collector.RecordGraphExtractorDroppedUnknown(evt.Source, 1)
	})
	eventbus.SubscribeTyped[eventbus.GraphAdmissionWriteFailureEvent](bus, func(evt eventbus.GraphAdmissionWriteFailureEvent) {
		oc.collector.RecordGraphWriteFailure(int64(evt.BatchCount))
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

func TestEventContract_GraphObservabilityEvents_AggregateInCollector(t *testing.T) {
	t.Parallel()

	bus, collector := wireTestObservability(t)
	learningGroup := "learning"

	bus.Publish(eventbus.GraphAdmissionBatchEvent{
		Source:           "learning",
		ProducerGroup:    &learningGroup,
		ValidatorSource:  "ontology_registry",
		BatchCount:       1,
		KnownCount:       2,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	})
	bus.Publish(eventbus.GraphAdmissionBatchEvent{
		Source:           "content_saved_extractor",
		ValidatorSource:  "unavailable",
		BatchCount:       1,
		KnownCount:       0,
		UnknownCount:     0,
		UnvalidatedCount: 3,
	})
	bus.Publish(eventbus.GraphAdmissionUnmappedSourceEvent{
		RawSource:  "legacy_import",
		BatchCount: 1,
	})
	bus.Publish(eventbus.GraphExtractorDroppedUnknownEvent{
		Source: "content_saved_extractor",
	})
	bus.Publish(eventbus.GraphAdmissionWriteFailureEvent{
		BatchCount: 1,
	})

	snap := collector.Snapshot()
	assert.Equal(t, observability.GraphAdmissionBatchMetric{
		Source:           "learning",
		ProducerGroup:    &learningGroup,
		ValidatorSource:  "ontology_registry",
		BatchCount:       1,
		KnownCount:       2,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	}, snap.GraphAdmission["learning|learning|ontology_registry"])
	assert.Equal(t, observability.GraphAdmissionBatchMetric{
		Source:           "content_saved_extractor",
		ProducerGroup:    nil,
		ValidatorSource:  "unavailable",
		BatchCount:       1,
		KnownCount:       0,
		UnknownCount:     0,
		UnvalidatedCount: 3,
	}, snap.GraphAdmission["content_saved_extractor||unavailable"])
	assert.Equal(t, int64(1), snap.GraphAdmissionUnmappedSources["legacy_import"])
	assert.Equal(t, int64(1), snap.GraphExtractorDroppedUnknown["content_saved_extractor"])
	assert.Equal(t, int64(1), snap.GraphWriteFailureBatches)
}
