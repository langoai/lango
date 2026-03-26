package knowledge

import (
	"fmt"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
)

// MapKnowledgeCategory maps an LLM analysis type string to a valid knowledge category.
// Returns error for unrecognized types (case-sensitive, no fallback).
func MapKnowledgeCategory(analysisType string) (entknowledge.Category, error) {
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

// MapLearningCategory maps an LLM analysis type string to a valid learning category.
// Returns error for unrecognized types (case-sensitive, no fallback).
func MapLearningCategory(analysisType string) (entlearning.Category, error) {
	switch analysisType {
	case "correction":
		return entlearning.CategoryUserCorrection, nil
	case "pattern":
		return entlearning.CategoryGeneral, nil
	case "tool_error":
		return entlearning.CategoryToolError, nil
	case "provider_error":
		return entlearning.CategoryProviderError, nil
	case "timeout":
		return entlearning.CategoryTimeout, nil
	case "permission":
		return entlearning.CategoryPermission, nil
	default:
		return "", fmt.Errorf("unrecognized learning type: %q", analysisType)
	}
}
