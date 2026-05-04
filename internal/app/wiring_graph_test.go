package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
)

func ptrString(v string) *string {
	return &v
}

func TestObserveExtractedTriples_PublishesAdmissionAndPreservesOriginalTriples(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var got eventbus.GraphAdmissionBatchEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionBatchEvent) {
		got = evt
	})

	policy := graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{
		Validator: func(name string) bool {
			return name == graph.CausedBy
		},
	}, zap.NewNop().Sugar())

	triples := observeExtractedTriples(policy, bus, eventbus.TriplesExtractedEvent{
		Source: string(graph.AdmissionSourceConversationAnalysis),
		Triples: []eventbus.Triple{{
			Subject:   "a",
			Predicate: "invented_rel",
			Object:    "b",
		}},
	})

	require.Len(t, triples, 1)
	assert.Equal(t, "invented_rel", triples[0].Predicate)
	assert.Equal(t, eventbus.GraphAdmissionBatchEvent{
		Source:           string(graph.AdmissionSourceConversationAnalysis),
		ProducerGroup:    ptrString(string(graph.AdmissionProducerGroupLearning)),
		ValidatorSource:  string(graph.AdmissionValidatorSourceOntologyRegistry),
		BatchCount:       1,
		KnownCount:       0,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	}, got)
}

func TestObserveExtractedTriples_PublishesUnmappedAndPreservesOriginalTriples(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var got eventbus.GraphAdmissionUnmappedSourceEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionUnmappedSourceEvent) {
		got = evt
	})

	policy := graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	triples := observeExtractedTriples(policy, bus, eventbus.TriplesExtractedEvent{
		Source: "future_source",
		Triples: []eventbus.Triple{{
			Subject:   "a",
			Predicate: "invented_rel",
			Object:    "b",
		}},
	})

	require.Len(t, triples, 1)
	assert.Equal(t, "invented_rel", triples[0].Predicate)
	assert.Equal(t, eventbus.GraphAdmissionUnmappedSourceEvent{
		RawSource:  "future_source",
		BatchCount: 1,
	}, got)
}

func TestObserveExtractedTriples_ProducerMappingCoverage(t *testing.T) {
	t.Parallel()

	policy := graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{
		Validator: func(name string) bool {
			return name == graph.CausedBy
		},
	}, zap.NewNop().Sugar())

	testCases := []struct {
		name              string
		source            string
		expectedGroup     *string
		expectBatch       bool
		expectUnmappedRaw string
	}{
		{
			name:          "conversation_analysis maps to learning",
			source:        string(graph.AdmissionSourceConversationAnalysis),
			expectedGroup: ptrString(string(graph.AdmissionProducerGroupLearning)),
			expectBatch:   true,
		},
		{
			name:          "session_learning maps to learning",
			source:        string(graph.AdmissionSourceSessionLearning),
			expectedGroup: ptrString(string(graph.AdmissionProducerGroupLearning)),
			expectBatch:   true,
		},
		{
			name:          "learning maps to learning",
			source:        string(graph.AdmissionSourceLearning),
			expectedGroup: ptrString(string(graph.AdmissionProducerGroupLearning)),
			expectBatch:   true,
		},
		{
			name:          "proactive_librarian maps to librarian",
			source:        string(graph.AdmissionSourceProactiveLibrarian),
			expectedGroup: ptrString(string(graph.AdmissionProducerGroupLibrarian)),
			expectBatch:   true,
		},
		{
			name:              "fallback source emits unmapped",
			source:            "future_source",
			expectUnmappedRaw: "future_source",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bus := eventbus.New()
			var gotBatch *eventbus.GraphAdmissionBatchEvent
			var gotUnmapped *eventbus.GraphAdmissionUnmappedSourceEvent
			eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionBatchEvent) {
				event := evt
				gotBatch = &event
			})
			eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionUnmappedSourceEvent) {
				event := evt
				gotUnmapped = &event
			})

			triples := observeExtractedTriples(policy, bus, eventbus.TriplesExtractedEvent{
				Source: tc.source,
				Triples: []eventbus.Triple{{
					Subject:   "a",
					Predicate: graph.CausedBy,
					Object:    "b",
				}},
			})

			require.Len(t, triples, 1)
			assert.Equal(t, graph.CausedBy, triples[0].Predicate)

			if tc.expectBatch {
				require.NotNil(t, gotBatch)
				assert.Equal(t, tc.source, gotBatch.Source)
				assert.Equal(t, tc.expectedGroup, gotBatch.ProducerGroup)
				assert.Equal(t, string(graph.AdmissionValidatorSourceOntologyRegistry), gotBatch.ValidatorSource)
				assert.Equal(t, 1, gotBatch.BatchCount)
				assert.Equal(t, 1, gotBatch.KnownCount)
				assert.Equal(t, 0, gotBatch.UnknownCount)
				assert.Equal(t, 0, gotBatch.UnvalidatedCount)
				assert.Nil(t, gotUnmapped)
				return
			}

			assert.Nil(t, gotBatch)
			require.NotNil(t, gotUnmapped)
			assert.Equal(t, tc.expectUnmappedRaw, gotUnmapped.RawSource)
			assert.Equal(t, 1, gotUnmapped.BatchCount)
		})
	}
}

func TestObserveContentSavedTriples_PublishesAdmissionAndPreservesOriginalTriples(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var got eventbus.GraphAdmissionBatchEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionBatchEvent) {
		got = evt
	})

	policy := graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{
		Validator: func(name string) bool {
			return name == graph.CausedBy
		},
	}, zap.NewNop().Sugar())

	triples := observeContentSavedTriples(policy, bus, []graph.Triple{{
		Subject:   "a",
		Predicate: graph.CausedBy,
		Object:    "b",
	}})

	require.Len(t, triples, 1)
	assert.Equal(t, graph.CausedBy, triples[0].Predicate)
	assert.Equal(t, eventbus.GraphAdmissionBatchEvent{
		Source:           string(graph.AdmissionSourceContentSavedExtractor),
		ProducerGroup:    nil,
		ValidatorSource:  string(graph.AdmissionValidatorSourceOntologyRegistry),
		BatchCount:       1,
		KnownCount:       1,
		UnknownCount:     0,
		UnvalidatedCount: 0,
	}, got)
}
