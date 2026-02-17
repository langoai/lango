package librarian

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/memory"
)

const observationAnalysisPrompt = `You are a proactive knowledge librarian. Analyze the following conversation observations and:

1. EXTRACT knowledge: user preferences, domain facts, rules, definitions, important context
2. DETECT gaps: ambiguous topics, incomplete information, contradictions, or areas where clarification would be valuable

For each extraction, assign a confidence level:
- "high": clearly stated by user, definitive fact or explicit preference
- "medium": implied or partially stated, needs confirmation
- "low": speculative, inferred from context

For each gap, assign a priority:
- "high": blocks understanding or could lead to errors
- "medium": would improve knowledge quality
- "low": nice to know, not urgent

Output JSON:
{
  "extractions": [
    {
      "type": "preference|fact|rule|definition",
      "category": "domain category",
      "content": "the knowledge content",
      "confidence": "high|medium|low",
      "key": "unique_snake_case_key",
      "subject": "optional graph subject",
      "predicate": "optional graph predicate",
      "object": "optional graph object"
    }
  ],
  "gaps": [
    {
      "topic": "topic of the gap",
      "question": "natural question to ask the user",
      "context": "why this question matters",
      "priority": "high|medium|low"
    }
  ]
}

Rules:
- Only extract genuinely useful knowledge, not conversational filler
- Questions should be natural and conversational, not survey-like
- Skip knowledge that is too session-specific or ephemeral
- Prefer specific, actionable knowledge over vague observations`

// ObservationAnalyzer uses LLM to analyze observations and extract knowledge/gaps.
type ObservationAnalyzer struct {
	generator TextGenerator
	logger    *zap.SugaredLogger
}

// NewObservationAnalyzer creates a new observation analyzer.
func NewObservationAnalyzer(generator TextGenerator, logger *zap.SugaredLogger) *ObservationAnalyzer {
	return &ObservationAnalyzer{
		generator: generator,
		logger:    logger,
	}
}

// Analyze processes observations through LLM to extract knowledge and detect gaps.
func (a *ObservationAnalyzer) Analyze(ctx context.Context, observations []memory.Observation) (*AnalysisOutput, error) {
	if len(observations) == 0 {
		return &AnalysisOutput{}, nil
	}

	// Build observation content for LLM.
	var content strings.Builder
	for i, obs := range observations {
		content.WriteString(fmt.Sprintf("--- Observation %d ---\n%s\n\n", i+1, obs.Content))
	}

	raw, err := a.generator.GenerateText(ctx, observationAnalysisPrompt, content.String())
	if err != nil {
		return nil, fmt.Errorf("analyze observations: %w", err)
	}

	output, err := parseAnalysisOutput(raw)
	if err != nil {
		a.logger.Warnw("parse analysis output", "error", err, "raw", raw)
		return &AnalysisOutput{}, nil
	}

	return output, nil
}
