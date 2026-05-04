package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

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
	assert.Equal(t, eventbus.GraphAdmissionBatchEvent{
		Source:           string(AdmissionSourceConversationAnalysis),
		ProducerGroup:    string(AdmissionProducerGroupLearning),
		ValidatorSource:  string(AdmissionValidatorSourceOntologyRegistry),
		BatchCount:       1,
		KnownCount:       1,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	}, result.Event)
}

func TestAdmissionPolicy_ObserveBatch_UsesUnvalidatedWhenValidatorUnavailable(t *testing.T) {
	t.Parallel()

	policy := NewAdmissionPolicy(AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	result := policy.ObserveBatch(AdmissionBatch{
		Source: AdmissionSourceContentSavedExtractor,
		Triples: []Triple{
			{Subject: "a", Predicate: CausedBy, Object: "b"},
			{Subject: "c", Predicate: "invented_rel", Object: "d"},
		},
	})

	require.Len(t, result.Records, 2)
	assert.Equal(t, AdmissionDecisionUnvalidated, result.Records[0].Decision)
	assert.Equal(t, AdmissionDecisionUnvalidated, result.Records[1].Decision)
	assert.Equal(t, eventbus.GraphAdmissionBatchEvent{
		Source:           string(AdmissionSourceContentSavedExtractor),
		ProducerGroup:    "",
		ValidatorSource:  string(AdmissionValidatorSourceUnavailable),
		BatchCount:       1,
		KnownCount:       0,
		UnknownCount:     0,
		UnvalidatedCount: 2,
	}, result.Event)
}
