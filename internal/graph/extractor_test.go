package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestExtractor_ValidatorRejectsUnknown(t *testing.T) {
	validator := func(name string) bool {
		return name == CausedBy || name == RelatedTo
	}
	logger := zap.NewNop().Sugar()
	e := NewExtractor(nil, logger, WithPredicateValidator(validator))

	// Valid predicate
	assert.True(t, e.isValidPredicate(CausedBy))
	assert.True(t, e.isValidPredicate(RelatedTo))

	// Invalid predicate — rejected by ontology validator
	assert.False(t, e.isValidPredicate("invented_rel"))
	assert.False(t, e.isValidPredicate(SimilarTo)) // not in validator's set
}

func TestExtractor_NoValidatorUsesHardcodedFallback(t *testing.T) {
	logger := zap.NewNop().Sugar()
	e := NewExtractor(nil, logger) // no WithPredicateValidator

	// All 9 hardcoded predicates accepted
	for _, p := range []string{RelatedTo, CausedBy, ResolvedBy, Follows, SimilarTo, Contains, InSession, ReflectsOn, LearnedFrom} {
		assert.True(t, e.isValidPredicate(p), "expected %q to be valid", p)
	}

	// Unknown rejected
	assert.False(t, e.isValidPredicate("made_up"))
}

func TestExtractor_ParseResponseRejectsInvalidPredicate(t *testing.T) {
	validator := func(name string) bool {
		return name == CausedBy
	}
	logger := zap.NewNop().Sugar()
	e := NewExtractor(nil, logger, WithPredicateValidator(validator))

	response := "a|caused_by|b\nc|fake_rel|d\ne|caused_by|f"
	triples := e.parseResponse(response, "test-source")

	assert.Len(t, triples, 2)
	assert.Equal(t, "a", triples[0].Subject)
	assert.Equal(t, "e", triples[1].Subject)
}

func TestExtractor_EmitDroppedUnknownObservation(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	var got []DroppedUnknownPredicateEvent
	e := NewExtractor(nil, logger,
		WithPredicateValidator(func(name string) bool {
			return name == CausedBy
		}),
		WithDroppedUnknownObserver(func(evt DroppedUnknownPredicateEvent) {
			got = append(got, evt)
		}),
	)

	triples := e.parseResponse("a|invented_rel|b", "src")

	assert.Len(t, triples, 0)
	require.Len(t, got, 1)
	assert.Equal(t, DroppedUnknownPredicateEvent{
		Source:    string(AdmissionSourceContentSavedExtractor),
		SourceID:  "src",
		Subject:   "a",
		Predicate: "invented_rel",
		Object:    "b",
	}, got[0])
}
