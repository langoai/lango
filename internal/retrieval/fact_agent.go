package retrieval

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/knowledge"
)

// FactSearchAgent searches knowledge, learnings, and external references.
type FactSearchAgent struct {
	source FactSearchSource
}

// NewFactSearchAgent creates a new FactSearchAgent backed by the given source.
func NewFactSearchAgent(source FactSearchSource) *FactSearchAgent {
	return &FactSearchAgent{source: source}
}

// Name returns the agent identifier.
func (a *FactSearchAgent) Name() string { return "fact-search" }

// Layers returns the context layers this agent produces.
func (a *FactSearchAgent) Layers() []knowledge.ContextLayer {
	return []knowledge.ContextLayer{
		knowledge.LayerUserKnowledge,
		knowledge.LayerAgentLearnings,
		knowledge.LayerExternalKnowledge,
	}
}

// Search retrieves findings from knowledge, learnings, and external references.
func (a *FactSearchAgent) Search(ctx context.Context, query string, limit int) ([]Finding, error) {
	if limit <= 0 {
		limit = 10
	}

	var findings []Finding

	knowledgeEntries, err := a.source.SearchKnowledgeScored(ctx, query, "", limit)
	if err != nil {
		return nil, fmt.Errorf("fact search knowledge: %w", err)
	}
	for _, k := range knowledgeEntries {
		findings = append(findings, Finding{
			Key:          k.Entry.Key,
			Content:      k.Entry.Content,
			Score:        k.Score,
			Category:     string(k.Entry.Category),
			SearchSource: k.SearchSource,
			Agent:        a.Name(),
			Layer:        knowledge.LayerUserKnowledge,
			Source:       k.Entry.Source,
			Tags:         k.Entry.Tags,
			Version:      k.Entry.Version,
			UpdatedAt:    k.Entry.UpdatedAt,
		})
	}

	learnings, err := a.source.SearchLearningsScored(ctx, query, "", limit)
	if err != nil {
		return nil, fmt.Errorf("fact search learnings: %w", err)
	}
	for _, l := range learnings {
		content := l.Entry.Trigger
		if l.Entry.Fix != "" {
			content = fmt.Sprintf("When '%s' occurs: %s", l.Entry.Trigger, l.Entry.Fix)
		}
		findings = append(findings, Finding{
			Key:          l.Entry.Trigger,
			Content:      content,
			Score:        l.Score,
			Category:     string(l.Entry.Category),
			SearchSource: l.SearchSource,
			Agent:        a.Name(),
			Layer:        knowledge.LayerAgentLearnings,
		})
	}

	refs, err := a.source.SearchExternalRefs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fact search external refs: %w", err)
	}
	for _, r := range refs {
		findings = append(findings, Finding{
			Key:          r.Name,
			Content:      r.Summary,
			Score:        0,
			Category:     r.RefType,
			SearchSource: "like",
			Agent:        a.Name(),
			Layer:        knowledge.LayerExternalKnowledge,
		})
	}

	return findings, nil
}
