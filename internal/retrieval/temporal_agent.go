package retrieval

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/langoai/lango/internal/knowledge"
)

// TemporalSearchSource is the narrow interface for recency-based search.
// Satisfied by *knowledge.Store.
type TemporalSearchSource interface {
	SearchRecentKnowledge(ctx context.Context, query string, limit int) ([]knowledge.KnowledgeEntry, error)
}

// Compile-time interface compliance check.
var _ RetrievalAgent = (*TemporalSearchAgent)(nil)

// maxAgeHours defines the recency decay window. Entries updated within this
// many hours score > 0; older entries score 0.
const maxAgeHours = 168 // 1 week

// TemporalSearchAgent retrieves recently-updated knowledge entries and scores
// them by recency. It complements FactSearchAgent (keyword) by surfacing items
// based on freshness rather than textual relevance.
//
// v1 scope: LayerUserKnowledge only. Learnings lack version chains and are excluded.
type TemporalSearchAgent struct {
	source TemporalSearchSource
	now    func() time.Time // injectable for testing
}

// NewTemporalSearchAgent creates a TemporalSearchAgent backed by the given source.
func NewTemporalSearchAgent(source TemporalSearchSource) *TemporalSearchAgent {
	return &TemporalSearchAgent{
		source: source,
		now:    time.Now,
	}
}

// Name returns the agent identifier.
func (a *TemporalSearchAgent) Name() string { return "temporal-search" }

// Layers returns the context layers this agent produces.
func (a *TemporalSearchAgent) Layers() []knowledge.ContextLayer {
	return []knowledge.ContextLayer{
		knowledge.LayerUserKnowledge,
	}
}

// Search retrieves recently-updated knowledge entries and scores by recency.
func (a *TemporalSearchAgent) Search(ctx context.Context, query string, limit int) ([]Finding, error) {
	if limit <= 0 {
		limit = 10
	}

	entries, err := a.source.SearchRecentKnowledge(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("temporal search: %w", err)
	}

	now := a.now()
	findings := make([]Finding, 0, len(entries))
	for _, e := range entries {
		score := recencyScore(e.UpdatedAt, now)
		findings = append(findings, Finding{
			Key:          e.Key,
			Content:      enrichTemporalContent(e.Content, e.Version, e.UpdatedAt, now),
			Score:        score,
			Category:     string(e.Category),
			SearchSource: "temporal",
			Agent:        a.Name(),
			Layer:        knowledge.LayerUserKnowledge,
			Source:       e.Source,
			Tags:         e.Tags,
			Version:      e.Version,
			UpdatedAt:    e.UpdatedAt,
		})
	}

	return findings, nil
}

// recencyScore computes a 0-1 score based on how recently the entry was updated.
// Score = max(0, 1.0 - hoursSinceUpdate / maxAgeHours).
// Recently updated entries score ~1.0, entries older than maxAgeHours score 0.0.
func recencyScore(updatedAt time.Time, now time.Time) float64 {
	hours := now.Sub(updatedAt).Hours()
	return math.Max(0, 1.0-hours/maxAgeHours)
}

// enrichTemporalContent prepends version and recency metadata to content.
func enrichTemporalContent(content string, version int, updatedAt time.Time, now time.Time) string {
	age := now.Sub(updatedAt)
	var ageStr string
	switch {
	case age < time.Minute:
		ageStr = "just now"
	case age < time.Hour:
		ageStr = fmt.Sprintf("%dm ago", int(age.Minutes()))
	case age < 24*time.Hour:
		ageStr = fmt.Sprintf("%dh ago", int(age.Hours()))
	default:
		ageStr = fmt.Sprintf("%dd ago", int(age.Hours()/24))
	}
	return fmt.Sprintf("[v%d | updated %s] %s", version, ageStr, content)
}
