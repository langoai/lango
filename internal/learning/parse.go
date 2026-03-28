package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/types"
)

// analysisResult is the expected structure from LLM analysis output.
type analysisResult struct {
	Type       string           `json:"type"`                // rule, definition, preference, fact, pattern, correction
	Category   string           `json:"category"`            // domain-specific category
	Content    string           `json:"content"`             // the extracted knowledge
	Confidence types.Confidence `json:"confidence"`          // low, medium, high
	Subject    string           `json:"subject,omitempty"`   // optional graph subject
	Predicate  string           `json:"predicate,omitempty"` // optional graph predicate
	Object     string           `json:"object,omitempty"`    // optional graph object
	Temporal   string           `json:"temporal,omitempty"`  // evergreen or current_state
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
func mapLearningCategory(analysisType string) (entlearning.Category, error) {
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

// saveResultParams holds the varying parts for saveAnalysisResult.
type saveResultParams struct {
	KeyPrefix     string // knowledge key prefix (e.g. "conv", "session")
	TriggerPrefix string // learning trigger prefix (e.g. "conversation", "session")
	Source        string // knowledge source label
}

// saveAnalysisResult persists an analysisResult as knowledge (all types) and
// additionally as a learning entry for pattern/correction (backward compat).
func saveAnalysisResult(
	ctx context.Context,
	store *knowledge.Store,
	bus *eventbus.Bus,
	logger *zap.SugaredLogger,
	sessionKey string,
	r analysisResult,
	p saveResultParams,
) {
	// ALL types → save as knowledge.
	cat, err := mapKnowledgeCategory(r.Type)
	if err != nil {
		logger.Debugw("skip knowledge: unknown type", "type", r.Type, "error", err)
		return
	}
	key := fmt.Sprintf("%s:%s:%s", p.KeyPrefix, sessionKey, sanitizeForNode(r.Content[:min(len(r.Content), 32)]))
	entry := knowledge.KnowledgeEntry{
		Key:      key,
		Category: cat,
		Content:  r.Content,
		Source:   p.Source,
	}
	if r.Temporal != "" {
		entry.Tags = append(entry.Tags, "temporal:"+r.Temporal)
	}
	if err := store.SaveKnowledge(ctx, sessionKey, entry); err != nil {
		logger.Debugw("save knowledge from analysis", "error", err)
	}

	// ADDITIONALLY for pattern/correction → also save as learning (backward compat).
	if r.Type == "pattern" || r.Type == "correction" {
		lCat, lErr := mapLearningCategory(r.Type)
		if lErr != nil {
			logger.Debugw("skip learning: unknown type", "type", r.Type, "error", lErr)
		} else {
			lEntry := knowledge.LearningEntry{
				Trigger:   fmt.Sprintf("%s:%s", p.TriggerPrefix, r.Category),
				Diagnosis: r.Content,
				Category:  lCat,
			}
			if r.Type == "correction" {
				lEntry.Fix = r.Content
				lEntry.Category = entlearning.CategoryUserCorrection
			}
			if err := store.SaveLearning(ctx, sessionKey, lEntry); err != nil {
				logger.Debugw("save learning from analysis", "error", err)
			}
		}
	}

	// Emit graph triples if provided.
	if bus != nil && r.Subject != "" && r.Predicate != "" && r.Object != "" {
		bus.Publish(eventbus.TriplesExtractedEvent{
			Triples: []eventbus.Triple{{
				Subject:   r.Subject,
				Predicate: r.Predicate,
				Object:    r.Object,
			}},
			Source: p.Source,
		})
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
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
