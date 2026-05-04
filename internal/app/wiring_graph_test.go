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
