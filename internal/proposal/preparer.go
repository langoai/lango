package proposal

import (
	"fmt"
	"strings"
	"time"
)

const learningProposalSourceKind = "proposed_learning"

// LearningSuggestionSource is the source-native input for the first Wave 3
// proposal preparation slice.
type LearningSuggestionSource struct {
	SessionKey   string
	SuggestionID string
	Pattern      string
	ProposedRule string
	Confidence   float64
	Rationale    string
	ExpiresAt    time.Time
}

// Preparer produces deterministic prepared briefs from source-native inputs.
type Preparer interface {
	PrepareLearningSuggestion(source LearningSuggestionSource) (PreparedBrief, error)
}

// DeterministicPreparer implements the Wave 3 low-risk preparation contract.
// It only formats already-available source fields and performs no external work.
type DeterministicPreparer struct{}

// NewDeterministicPreparer returns the source-native preparer for the learning-suggestion slice.
func NewDeterministicPreparer() *DeterministicPreparer {
	return &DeterministicPreparer{}
}

// PrepareLearningSuggestion formats a deterministic prepared brief from an existing
// learning suggestion source. No external state is mutated and no broad heuristics
// are used.
func (p *DeterministicPreparer) PrepareLearningSuggestion(source LearningSuggestionSource) (PreparedBrief, error) {
	return PreparedBrief{
		SourceSummary:             learningSuggestionSourceSummary(source),
		Reason:                    learningSuggestionReason(source),
		SuggestedAcceptanceEffect: learningSuggestionAcceptanceEffect(source),
		SupportingEvidence:        learningSuggestionEvidence(source),
	}, nil
}

func learningSuggestionTitle(source LearningSuggestionSource) string {
	if rule := strings.TrimSpace(source.ProposedRule); rule != "" {
		return "Apply learning rule: " + rule
	}
	if pattern := strings.TrimSpace(source.Pattern); pattern != "" {
		return "Review learning suggestion: " + pattern
	}
	return "Review learning suggestion"
}

func learningSuggestionSourceSummary(source LearningSuggestionSource) string {
	id := learningSuggestionLabel(source)
	rule := strings.TrimSpace(source.ProposedRule)
	pattern := strings.TrimSpace(source.Pattern)

	switch {
	case rule != "" && pattern != "":
		return fmt.Sprintf(`Learning suggestion %q proposes rule %q for pattern %q.`, id, rule, pattern)
	case rule != "":
		return fmt.Sprintf(`Learning suggestion %q proposes rule %q.`, id, rule)
	case pattern != "":
		return fmt.Sprintf(`Learning suggestion %q highlights pattern %q.`, id, pattern)
	default:
		return fmt.Sprintf(`Learning suggestion %q is available for review.`, id)
	}
}

func learningSuggestionReason(source LearningSuggestionSource) string {
	if reason := strings.TrimSpace(source.Rationale); reason != "" {
		return reason
	}
	return "No explicit rationale was provided with this learning suggestion."
}

func learningSuggestionAcceptanceEffect(source LearningSuggestionSource) string {
	if rule := strings.TrimSpace(source.ProposedRule); rule != "" {
		return fmt.Sprintf(`Accepting will create a durable mission to apply learning rule %q.`, rule)
	}
	if pattern := strings.TrimSpace(source.Pattern); pattern != "" {
		return fmt.Sprintf(`Accepting will create a durable mission to investigate pattern %q.`, pattern)
	}
	return "Accepting will create a durable mission to review this learning suggestion."
}

func learningSuggestionEvidence(source LearningSuggestionSource) []string {
	evidence := make([]string, 0, 4)
	if id := strings.TrimSpace(source.SuggestionID); id != "" {
		evidence = append(evidence, "Suggestion ID: "+id)
	}
	if pattern := strings.TrimSpace(source.Pattern); pattern != "" {
		evidence = append(evidence, "Pattern: "+pattern)
	}
	if rule := strings.TrimSpace(source.ProposedRule); rule != "" {
		evidence = append(evidence, "Proposed rule: "+rule)
	}
	if source.Confidence > 0 {
		evidence = append(evidence, fmt.Sprintf("Confidence: %.2f", source.Confidence))
	}
	return evidence
}

func learningSuggestionLabel(source LearningSuggestionSource) string {
	if id := strings.TrimSpace(source.SuggestionID); id != "" {
		return id
	}
	return "unspecified"
}
