package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

func strPtr(v string) *string {
	return &v
}

func TestAdmissionPolicy_ObserveBatch_ClassifiesKnownUnknown(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{
		Validator: func(name string) bool {
			return name == CausedBy
		},
	}, zap.NewNop().Sugar())

	result := policy.ObserveBatch(AdmissionBatch{
		Source:        AdmissionSourceConversationAnalysis,
		ProducerGroup: AdmissionProducerGroupLearning,
		Triples: []Triple{
			{Subject: "known", Predicate: CausedBy, Object: "accepted"},
			{Subject: "unknown", Predicate: "invented_rel", Object: "rejected"},
		},
	})

	require.Len(t, result.Records, 2)
	assert.Equal(t, AdmissionDecisionKnown, result.Records[0].Decision)
	assert.Equal(t, AdmissionDecisionUnknown, result.Records[1].Decision)
	assert.Equal(t, []Triple{
		{Subject: "known", Predicate: CausedBy, Object: "accepted"},
		{Subject: "unknown", Predicate: "invented_rel", Object: "rejected"},
	}, result.Forwarded)
	require.NotNil(t, result.Event)
	assert.Equal(t, eventbus.GraphAdmissionBatchEvent{
		Source:           string(AdmissionSourceConversationAnalysis),
		ProducerGroup:    strPtr(string(AdmissionProducerGroupLearning)),
		ValidatorSource:  string(AdmissionValidatorSourceOntologyRegistry),
		BatchCount:       1,
		KnownCount:       1,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	}, *result.Event)
}

func TestAdmissionPolicy_ObserveBatch_UsesUnvalidatedWhenValidatorUnavailable(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	result := policy.ObserveBatch(AdmissionBatch{
		SourceKind: AdmissionSourceKindSynthetic,
		Source:     AdmissionSourceContentSavedExtractor,
		Triples: []Triple{
			{Subject: "a", Predicate: CausedBy, Object: "b"},
			{Subject: "c", Predicate: "invented_rel", Object: "d"},
		},
	})

	require.Len(t, result.Records, 2)
	assert.Equal(t, AdmissionDecisionUnvalidated, result.Records[0].Decision)
	assert.Equal(t, AdmissionDecisionUnvalidated, result.Records[1].Decision)
	require.NotNil(t, result.Event)
	assert.Equal(t, eventbus.GraphAdmissionBatchEvent{
		Source:           string(AdmissionSourceContentSavedExtractor),
		ProducerGroup:    nil,
		ValidatorSource:  string(AdmissionValidatorSourceUnavailable),
		BatchCount:       1,
		KnownCount:       0,
		UnknownCount:     0,
		UnvalidatedCount: 2,
	}, *result.Event)
}

func TestAdmissionPolicy_ObserveBatch_SkipsUnsupportedSource(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{
		Validator: func(name string) bool {
			return name == CausedBy
		},
	}, zap.NewNop().Sugar())

	triples := []Triple{
		{Subject: "unsupported", Predicate: "invented_rel", Object: "forwarded"},
	}

	result := policy.ObserveBatch(AdmissionBatch{
		SourceKind: AdmissionSourceKindEventBus,
		Source:     AdmissionSource("new_source"),
		Triples:    triples,
	})

	assert.Empty(t, result.Records)
	assert.Equal(t, triples, result.Forwarded)
	assert.Nil(t, result.Event)
	require.NotNil(t, result.UnmappedEvent)
	assert.Equal(t, eventbus.GraphAdmissionUnmappedSourceEvent{
		RawSource:  "new_source",
		BatchCount: 1,
	}, *result.UnmappedEvent)
}

func TestAdmissionPolicy_ObserveBatch_ZeroValueSourceKindStillEmitsUnmapped(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	result := policy.ObserveBatch(AdmissionBatch{
		Source: AdmissionSource("unmapped_runtime_label"),
		Triples: []Triple{
			{Subject: "a", Predicate: "invented_rel", Object: "b"},
		},
	})

	assert.Nil(t, result.Event)
	require.NotNil(t, result.UnmappedEvent)
	assert.Equal(t, eventbus.GraphAdmissionUnmappedSourceEvent{
		RawSource:  "unmapped_runtime_label",
		BatchCount: 1,
	}, *result.UnmappedEvent)
}

func TestAdmissionPolicy_ObserveBatch_TreatsEventBusContentSavedExtractorAsUnmapped(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	result := policy.ObserveBatch(AdmissionBatch{
		Source: AdmissionSourceContentSavedExtractor,
		Triples: []Triple{
			{Subject: "a", Predicate: CausedBy, Object: "b"},
		},
	})

	assert.Nil(t, result.Event)
	require.NotNil(t, result.UnmappedEvent)
	assert.Equal(t, eventbus.GraphAdmissionUnmappedSourceEvent{
		RawSource:  string(AdmissionSourceContentSavedExtractor),
		BatchCount: 1,
	}, *result.UnmappedEvent)
}

func TestAdmissionPolicy_ObserveBatch_CanonicalizesKnownEventBusSourceKind(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	result := policy.ObserveBatch(AdmissionBatch{
		SourceKind: AdmissionSourceKindSynthetic,
		Source:     AdmissionSourceConversationAnalysis,
		Triples: []Triple{
			{Subject: "a", Predicate: CausedBy, Object: "b"},
		},
	})

	require.NotNil(t, result.Event)
	require.NotNil(t, result.Event.ProducerGroup)
	assert.Equal(t, string(AdmissionProducerGroupLearning), *result.Event.ProducerGroup)
	assert.Nil(t, result.UnmappedEvent)
}

func TestAdmissionPolicy_ObserveBatch_DoesNotEmitUnmappedForUnknownSyntheticLabel(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	triples := []Triple{
		{Subject: "synthetic", Predicate: "invented_rel", Object: "forwarded"},
	}

	result := policy.ObserveBatch(AdmissionBatch{
		SourceKind: AdmissionSourceKindSynthetic,
		Source:     AdmissionSource("future_synthetic_label"),
		Triples:    triples,
	})

	assert.Empty(t, result.Records)
	assert.Equal(t, triples, result.Forwarded)
	assert.Nil(t, result.Event)
	assert.Nil(t, result.UnmappedEvent)
}

func TestAdmissionPolicy_ObserveBatch_NormalizesContentSavedProducerGroup(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	result := policy.ObserveBatch(AdmissionBatch{
		SourceKind:    AdmissionSourceKindSynthetic,
		Source:        AdmissionSourceContentSavedExtractor,
		ProducerGroup: AdmissionProducerGroupLearning,
		Triples: []Triple{
			{Subject: "a", Predicate: CausedBy, Object: "b"},
		},
	})

	require.NotNil(t, result.Event)
	assert.Nil(t, result.Event.ProducerGroup)
}

func TestAdmissionPolicy_ObserveBatch_CanonicalizesProducerGroupFromSource(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	testCases := []struct {
		name          string
		source        AdmissionSource
		inputGroup    AdmissionProducerGroup
		expectedGroup AdmissionProducerGroup
	}{
		{
			name:          "conversation analysis maps to learning",
			source:        AdmissionSourceConversationAnalysis,
			inputGroup:    AdmissionProducerGroupLibrarian,
			expectedGroup: AdmissionProducerGroupLearning,
		},
		{
			name:          "session learning maps to learning",
			source:        AdmissionSourceSessionLearning,
			inputGroup:    AdmissionProducerGroupLibrarian,
			expectedGroup: AdmissionProducerGroupLearning,
		},
		{
			name:          "learning maps to learning",
			source:        AdmissionSourceLearning,
			inputGroup:    AdmissionProducerGroupLibrarian,
			expectedGroup: AdmissionProducerGroupLearning,
		},
		{
			name:          "proactive librarian maps to librarian",
			source:        AdmissionSourceProactiveLibrarian,
			inputGroup:    AdmissionProducerGroupLearning,
			expectedGroup: AdmissionProducerGroupLibrarian,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := policy.ObserveBatch(AdmissionBatch{
				Source:        tc.source,
				ProducerGroup: tc.inputGroup,
				Triples: []Triple{
					{Subject: "a", Predicate: CausedBy, Object: "b"},
				},
			})

			require.NotNil(t, result.Event)
			require.NotNil(t, result.Event.ProducerGroup)
			assert.Equal(t, string(tc.expectedGroup), *result.Event.ProducerGroup)
		})
	}
}

func TestAdmissionIdentifiers_AreStableForTask4Consumers(t *testing.T) {
	t.Parallel()

	assert.True(t, IsSupportedAdmissionSource(AdmissionSourceConversationAnalysis))
	assert.True(t, IsSupportedAdmissionSource(AdmissionSourceSessionLearning))
	assert.True(t, IsSupportedAdmissionSource(AdmissionSourceLearning))
	assert.True(t, IsSupportedAdmissionSource(AdmissionSourceProactiveLibrarian))
	assert.False(t, IsSupportedAdmissionSource(AdmissionSourceContentSavedExtractor))
	assert.False(t, IsSupportedAdmissionSource(AdmissionSource("new_source")))

	require.Equal(t, strPtr(string(AdmissionProducerGroupLearning)), CanonicalAdmissionProducerGroup(AdmissionSourceConversationAnalysis))
	require.Equal(t, strPtr(string(AdmissionProducerGroupLearning)), CanonicalAdmissionProducerGroup(AdmissionSourceSessionLearning))
	require.Equal(t, strPtr(string(AdmissionProducerGroupLearning)), CanonicalAdmissionProducerGroup(AdmissionSourceLearning))
	require.Equal(t, strPtr(string(AdmissionProducerGroupLibrarian)), CanonicalAdmissionProducerGroup(AdmissionSourceProactiveLibrarian))
	assert.Nil(t, CanonicalAdmissionProducerGroup(AdmissionSourceContentSavedExtractor))

	sourceKind, ok := ObservedAdmissionSourceKind(AdmissionSourceConversationAnalysis, "")
	require.True(t, ok)
	assert.Equal(t, AdmissionSourceKindEventBus, sourceKind)

	sourceKind, ok = ObservedAdmissionSourceKind(AdmissionSourceContentSavedExtractor, AdmissionSourceKindSynthetic)
	require.True(t, ok)
	assert.Equal(t, AdmissionSourceKindSynthetic, sourceKind)

	_, ok = ObservedAdmissionSourceKind(AdmissionSourceContentSavedExtractor, "")
	assert.False(t, ok)

	_, ok = ObservedAdmissionSourceKind(AdmissionSource("new_source"), "")
	assert.False(t, ok)
}
