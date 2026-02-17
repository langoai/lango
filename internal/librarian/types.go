package librarian

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ObservationKnowledge represents knowledge extracted from conversation observations.
type ObservationKnowledge struct {
	Type       string `json:"type"`       // preference, fact, rule, definition
	Category   string `json:"category"`   // domain-specific category
	Content    string `json:"content"`    // extracted knowledge content
	Confidence string `json:"confidence"` // high, medium, low
	Key        string `json:"key"`        // unique identifier for storage

	// Graph triple fields (optional).
	Subject   string `json:"subject,omitempty"`
	Predicate string `json:"predicate,omitempty"`
	Object    string `json:"object,omitempty"`
}

// KnowledgeGap represents a detected gap in knowledge that requires user clarification.
type KnowledgeGap struct {
	Topic       string `json:"topic"`
	Question    string `json:"question"`
	Context     string `json:"context,omitempty"`
	Priority    string `json:"priority"` // low, medium, high
	RelatedKeys []string `json:"relatedKeys,omitempty"`
}

// AnalysisOutput is the combined result from observation analysis.
type AnalysisOutput struct {
	Extractions []ObservationKnowledge `json:"extractions"`
	Gaps        []KnowledgeGap         `json:"gaps"`
}

// Inquiry represents a pending question to ask the user.
type Inquiry struct {
	ID                  uuid.UUID
	SessionKey          string
	Topic               string
	Question            string
	Context             string
	Priority            string // low, medium, high
	Status              string // pending, resolved, dismissed
	Answer              string
	KnowledgeKey        string
	SourceObservationID string
	CreatedAt           time.Time
	ResolvedAt          *time.Time
}

// TextGenerator abstracts LLM text generation for the librarian package.
type TextGenerator interface {
	GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// GraphCallback is an optional hook for saving graph triples.
type GraphCallback func(triples []Triple)

// Triple mirrors graph.Triple to avoid import cycles.
type Triple struct {
	Subject   string
	Predicate string
	Object    string
	Metadata  map[string]string
}
