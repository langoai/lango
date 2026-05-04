package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
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

type failingBufferStore struct {
	err error
}

func (s *failingBufferStore) AddTriple(context.Context, graph.Triple) error    { return s.err }
func (s *failingBufferStore) AddTriples(context.Context, []graph.Triple) error { return s.err }
func (s *failingBufferStore) RemoveTriple(context.Context, graph.Triple) error { return s.err }
func (s *failingBufferStore) QueryBySubject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingBufferStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingBufferStore) QueryBySubjectPredicate(context.Context, string, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingBufferStore) Traverse(context.Context, string, int, []string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingBufferStore) Count(context.Context) (int, error) { return 0, nil }
func (s *failingBufferStore) PredicateStats(context.Context) (map[string]int, error) {
	return nil, nil
}
func (s *failingBufferStore) AllTriples(context.Context) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingBufferStore) ClearAll(context.Context) error { return nil }
func (s *failingBufferStore) Close() error                   { return nil }

func TestWireGraphWriteFailureBaselineObserver_ObserveModeToggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mode          string
		wantPublished bool
	}{
		{
			name:          "observe mode publishes write failure baseline",
			mode:          config.OntologyAdmissionModeObserve,
			wantPublished: true,
		},
		{
			name:          "off mode suppresses write failure baseline",
			mode:          config.OntologyAdmissionModeOff,
			wantPublished: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.DefaultConfig()
			cfg.Ontology.Governance.AdmissionMode = tt.mode

			bus := eventbus.New()
			buffer := graph.NewGraphBuffer(&failingBufferStore{err: errors.New("boom")}, zap.NewNop().Sugar())
			wireGraphWriteFailureBaselineObserver(cfg, buffer, bus)

			var got []eventbus.GraphAdmissionWriteFailureEvent
			eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionWriteFailureEvent) {
				got = append(got, evt)
			})

			var wg sync.WaitGroup
			buffer.Start(&wg)
			buffer.Enqueue(graph.GraphRequest{
				Triples: []graph.Triple{{Subject: "a", Predicate: graph.Contains, Object: "b"}},
			})
			buffer.Stop()
			wg.Wait()

			if tt.wantPublished {
				require.Len(t, got, 1)
				assert.Equal(t, eventbus.GraphAdmissionWriteFailureEvent{BatchCount: 1}, got[0])
				return
			}
			assert.Empty(t, got)
		})
	}
}
