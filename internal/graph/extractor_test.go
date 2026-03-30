package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
