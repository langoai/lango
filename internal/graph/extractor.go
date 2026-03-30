package graph

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/llm"
)

// Extractor uses an LLM to extract entities and relationships from text.
type Extractor struct {
	generator llm.TextGenerator
	validator PredicateValidatorFunc
	logger    *zap.SugaredLogger
}

// ExtractorOption configures optional Extractor behavior.
type ExtractorOption func(*Extractor)

// WithPredicateValidator injects an ontology-backed predicate validator.
// When set, extracted predicates are validated against the registry.
// When not set, the hardcoded 9-predicate list is used as fallback.
func WithPredicateValidator(v PredicateValidatorFunc) ExtractorOption {
	return func(e *Extractor) { e.validator = v }
}

// NewExtractor creates a new LLM-based entity/relationship extractor.
func NewExtractor(generator llm.TextGenerator, logger *zap.SugaredLogger, opts ...ExtractorOption) *Extractor {
	e := &Extractor{
		generator: generator,
		logger:    logger,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

const extractionSystemPrompt = `You are an entity and relationship extraction system. Given text, extract entities and relationships as triples.

Output format (one triple per line):
SUBJECT|PREDICATE|OBJECT

Valid predicates: related_to, caused_by, resolved_by, follows, similar_to, contains

Rules:
- Extract factual relationships only
- Use concise entity names (lowercase, underscored)
- Skip trivial or obvious relationships
- Maximum 10 triples per extraction
- If no meaningful relationships found, output NONE

Example:
Input: "JWT token expired causing authentication failure. Fixed by implementing token refresh."
Output:
jwt_token_expiry|caused_by|authentication_failure
token_refresh|resolved_by|authentication_failure
jwt_token_expiry|related_to|token_refresh`

// Extract extracts triples from the given text content.
// The sourceID is used as context for provenance tracking.
func (e *Extractor) Extract(ctx context.Context, content, sourceID string) ([]Triple, error) {
	if content == "" {
		return nil, nil
	}

	userPrompt := fmt.Sprintf("Extract entities and relationships from:\n\n%s", content)

	response, err := e.generator.GenerateText(ctx, extractionSystemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("generate extraction: %w", err)
	}

	return e.parseResponse(response, sourceID), nil
}

// parseResponse parses the LLM response into triples.
func (e *Extractor) parseResponse(response, sourceID string) []Triple {
	response = strings.TrimSpace(response)
	if response == "" || response == "NONE" {
		return nil
	}

	lines := strings.Split(response, "\n")
	triples := make([]Triple, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "NONE" {
			continue
		}

		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			e.logger.Debugw("skip malformed triple", "line", line)
			continue
		}

		subject := strings.TrimSpace(parts[0])
		predicate := strings.TrimSpace(parts[1])
		object := strings.TrimSpace(parts[2])

		if subject == "" || predicate == "" || object == "" {
			continue
		}

		if !e.isValidPredicate(predicate) {
			e.logger.Warnw("rejected unknown predicate from extraction",
				"predicate", predicate,
				"subject", subject,
				"object", object,
				"source", sourceID,
			)
			continue
		}

		triples = append(triples, Triple{
			Subject:   subject,
			Predicate: predicate,
			Object:    object,
			Metadata: map[string]string{
				"source": sourceID,
			},
		})
	}

	return triples
}

// isValidPredicate validates using the ontology validator if set, otherwise hardcoded fallback.
func (e *Extractor) isValidPredicate(p string) bool {
	if e.validator != nil {
		return e.validator(p)
	}
	return defaultIsValidPredicate(p)
}

// defaultIsValidPredicate is the hardcoded fallback when no ontology validator is set.
func defaultIsValidPredicate(p string) bool {
	switch p {
	case RelatedTo, CausedBy, ResolvedBy, Follows, SimilarTo, Contains, InSession, ReflectsOn, LearnedFrom:
		return true
	}
	return false
}
