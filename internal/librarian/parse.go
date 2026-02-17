package librarian

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parseAnalysisOutput extracts structured analysis from LLM JSON response.
// Handles code fences and various JSON formats.
func parseAnalysisOutput(raw string) (*AnalysisOutput, error) {
	cleaned := stripCodeFence(raw)
	cleaned = strings.TrimSpace(cleaned)

	var output AnalysisOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return nil, fmt.Errorf("parse analysis output: %w", err)
	}
	return &output, nil
}

// parseAnswerMatches extracts answer matches from LLM JSON response.
func parseAnswerMatches(raw string) ([]answerMatch, error) {
	cleaned := stripCodeFence(raw)
	cleaned = strings.TrimSpace(cleaned)

	var matches []answerMatch
	if err := json.Unmarshal([]byte(cleaned), &matches); err != nil {
		return nil, fmt.Errorf("parse answer matches: %w", err)
	}
	return matches, nil
}

// answerMatch represents an LLM-detected match between a user message and a pending inquiry.
type answerMatch struct {
	InquiryID  string `json:"inquiry_id"`
	Answer     string `json:"answer"`
	Confidence string `json:"confidence"` // high, medium, low
	Knowledge  *matchedKnowledge `json:"knowledge,omitempty"`
}

// matchedKnowledge is the structured knowledge to save from a matched answer.
type matchedKnowledge struct {
	Key      string `json:"key"`
	Category string `json:"category"`
	Content  string `json:"content"`
}

// stripCodeFence removes markdown code fences from LLM output.
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
