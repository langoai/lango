package proposal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreparerLearningSuggestionYieldsStablePreparedBrief(t *testing.T) {
	t.Parallel()

	preparer := NewDeterministicPreparer()
	source := LearningSuggestionSource{
		SessionKey:   "sess-1",
		SuggestionID: "suggestion-1",
		Pattern:      "retry timeout",
		ProposedRule: "Use bounded retry with backoff",
		Confidence:   0.625,
		Rationale:    "Repeated timeout failures benefited from bounded retry.",
		ExpiresAt:    time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	}

	first, err := preparer.PrepareLearningSuggestion(source)
	require.NoError(t, err)
	second, err := preparer.PrepareLearningSuggestion(source)
	require.NoError(t, err)

	assert.Equal(t, first, second)
	assert.Equal(t,
		`Learning suggestion "suggestion-1" proposes rule "Use bounded retry with backoff" for pattern "retry timeout".`,
		first.SourceSummary,
	)
	assert.Equal(t, "Repeated timeout failures benefited from bounded retry.", first.Reason)
	assert.Equal(t,
		`Accepting will create a durable mission to apply learning rule "Use bounded retry with backoff".`,
		first.SuggestedAcceptanceEffect,
	)
	assert.Equal(t, []string{
		"Suggestion ID: suggestion-1",
		"Pattern: retry timeout",
		"Proposed rule: Use bounded retry with backoff",
		"Confidence: 0.62",
	}, first.SupportingEvidence)
}

func TestPreparerGracefullyDegradesPartialSourceFields(t *testing.T) {
	t.Parallel()

	preparer := NewDeterministicPreparer()
	source := LearningSuggestionSource{
		SessionKey:   "sess-1",
		SuggestionID: "suggestion-2",
		Rationale:    "",
	}

	brief, err := preparer.PrepareLearningSuggestion(source)
	require.NoError(t, err)

	assert.Equal(t,
		`Learning suggestion "suggestion-2" is available for review.`,
		brief.SourceSummary,
	)
	assert.Equal(t,
		"No explicit rationale was provided with this learning suggestion.",
		brief.Reason,
	)
	assert.Equal(t,
		"Accepting will create a durable mission to review this learning suggestion.",
		brief.SuggestedAcceptanceEffect,
	)
	assert.Equal(t, []string{"Suggestion ID: suggestion-2"}, brief.SupportingEvidence)
}

func TestPreparerDoesNotImplyExternalStateMutation(t *testing.T) {
	t.Parallel()

	preparer := NewDeterministicPreparer()
	source := LearningSuggestionSource{
		SessionKey:   "sess-2",
		SuggestionID: "suggestion-3",
		Pattern:      "planner drift",
		ProposedRule: "Re-anchor the plan before execution",
		Confidence:   0.4,
		Rationale:    "Planner output diverged from accepted scope.",
	}
	original := source

	_, err := preparer.PrepareLearningSuggestion(source)
	require.NoError(t, err)

	assert.Equal(t, original, source)
}
