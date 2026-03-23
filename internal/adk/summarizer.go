package adk

import (
	"strings"

	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

// Summarizer produces a summary string from a child session's messages.
type Summarizer interface {
	Summarize(messages []session.Message) (string, error)
}

// StructuredSummarizer extracts the last assistant response as the summary.
// This is the default zero-cost summarizer that avoids LLM calls.
type StructuredSummarizer struct{}

// Summarize returns the last assistant message content as the summary.
// If no assistant message is found, returns an empty string.
func (s *StructuredSummarizer) Summarize(messages []session.Message) (string, error) {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == types.RoleAssistant &&
			strings.TrimSpace(messages[i].Content) != "" {
			return messages[i].Content, nil
		}
	}
	return "", nil
}
