package learning

import (
	"encoding/json"
	"fmt"
	"strings"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/types"
)

// analysisResult is the expected structure from LLM analysis output.
type analysisResult struct {
	Type       string           `json:"type"`                // fact, pattern, correction, preference
	Category   string           `json:"category"`            // domain-specific category
	Content    string           `json:"content"`             // the extracted knowledge
	Confidence types.Confidence `json:"confidence"`          // low, medium, high
	Subject    string           `json:"subject,omitempty"`   // optional graph subject
	Predicate  string           `json:"predicate,omitempty"` // optional graph predicate
	Object     string           `json:"object,omitempty"`    // optional graph object
}

// parseAnalysisResponse extracts structured results from an LLM JSON response.
// Handles code fences, single objects, and arrays.
func parseAnalysisResponse(raw string) ([]analysisResult, error) {
	cleaned := stripCodeFence(raw)
	cleaned = strings.TrimSpace(cleaned)

	// Try array first.
	var results []analysisResult
	if err := json.Unmarshal([]byte(cleaned), &results); err == nil {
		return results, nil
	}

	// Try single object.
	var single analysisResult
	if err := json.Unmarshal([]byte(cleaned), &single); err == nil {
		return []analysisResult{single}, nil
	}

	return nil, fmt.Errorf("parse analysis response: invalid JSON")
}

// mapKnowledgeCategory maps LLM analysis type to a valid knowledge category.
func mapKnowledgeCategory(analysisType string) (entknowledge.Category, error) {
	switch analysisType {
	case "preference":
		return entknowledge.CategoryPreference, nil
	case "fact":
		return entknowledge.CategoryFact, nil
	case "rule":
		return entknowledge.CategoryRule, nil
	case "definition":
		return entknowledge.CategoryDefinition, nil
	case "pattern":
		return entknowledge.CategoryPattern, nil
	case "correction":
		return entknowledge.CategoryCorrection, nil
	default:
		return "", fmt.Errorf("unrecognized knowledge type: %q", analysisType)
	}
}

// mapLearningCategory maps LLM analysis type to a valid learning category.
func mapLearningCategory(analysisType string) entlearning.Category {
	switch analysisType {
	case "correction":
		return entlearning.CategoryUserCorrection
	case "pattern":
		return entlearning.CategoryGeneral
	case "tool_error":
		return entlearning.CategoryToolError
	case "provider_error":
		return entlearning.CategoryProviderError
	case "timeout":
		return entlearning.CategoryTimeout
	case "permission":
		return entlearning.CategoryPermission
	default:
		return entlearning.CategoryGeneral
	}
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
