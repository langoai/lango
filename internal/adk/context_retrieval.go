package adk

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/types"
	"google.golang.org/genai"
)

// extractLastUserMessage finds the last user message from the content history.
func extractLastUserMessage(contents []*genai.Content) string {
	for i := len(contents) - 1; i >= 0; i-- {
		c := contents[i]
		if c.Role == "user" {
			for _, p := range c.Parts {
				if p.Text != "" {
					return p.Text
				}
			}
		}
	}
	return ""
}

// assembleGraphRAGSection builds a combined section from vector search + graph expansion.
// budgetTokens controls item-level truncation: 0 = unlimited.
func (m *ContextAwareModelAdapter) assembleGraphRAGSection(ctx context.Context, query, sessionKey string, budgetTokens int) string {
	opts := graph.VectorRetrieveOptions{
		Collections: m.ragOpts.Collections,
		Limit:       m.ragOpts.Limit,
		SessionKey:  m.ragOpts.SessionKey,
		MaxDistance:  m.ragOpts.MaxDistance,
	}
	if sessionKey != "" {
		opts.SessionKey = sessionKey
	}
	result, err := m.graphRAG.Retrieve(ctx, query, opts)
	if err != nil {
		m.logger.Warnw("graph rag retrieval error", "error", err)
		return ""
	}

	// Item-level truncation: drop results from both vector and graph lists.
	if budgetTokens > 0 && result != nil {
		remaining := budgetTokens
		// Truncate vector results.
		for i, r := range result.VectorResults {
			itemTokens := types.EstimateTokens(r.Content) + types.EstimateTokens(fmt.Sprintf("\n### [%s] %s\n", r.Collection, r.SourceID))
			if remaining-itemTokens < 0 {
				result.VectorResults = result.VectorResults[:i]
				result.GraphResults = nil // no budget left for graph
				break
			}
			remaining -= itemTokens
		}
		// Truncate graph results with remaining budget.
		for i, g := range result.GraphResults {
			itemTokens := types.EstimateTokens(fmt.Sprintf("- **%s** (via %s from %s)\n", g.ID, g.Predicate, g.FromNode))
			if remaining-itemTokens < 0 {
				result.GraphResults = result.GraphResults[:i]
				break
			}
			remaining -= itemTokens
		}
	}

	return m.graphRAG.AssembleSection(result)
}

// assembleRAGSection builds a "Semantic Context" section from RAG retrieval results.
// budgetTokens controls item-level truncation: 0 = unlimited.
func (m *ContextAwareModelAdapter) assembleRAGSection(ctx context.Context, query, sessionKey string, budgetTokens int) string {
	opts := m.ragOpts
	if sessionKey != "" {
		opts.SessionKey = sessionKey
	}
	results, err := m.ragService.Retrieve(ctx, query, opts)
	if err != nil {
		m.logger.Warnw("rag retrieval error", "error", err)
		return ""
	}
	if len(results) == 0 {
		return ""
	}

	// Item-level truncation: drop lowest-rank results until within budget.
	if budgetTokens > 0 {
		headerTokens := types.EstimateTokens("## Semantic Context (RAG)\n")
		remaining := budgetTokens - headerTokens
		kept := 0
		for i, r := range results {
			itemTokens := types.EstimateTokens(r.Content) + types.EstimateTokens(fmt.Sprintf("\n### [%s] %s\n", r.Collection, r.SourceID))
			if remaining-itemTokens < 0 {
				results = results[:i]
				break
			}
			remaining -= itemTokens
			kept++
		}
		if kept == 0 {
			return ""
		}
	}

	var b strings.Builder
	b.WriteString("## Semantic Context (RAG)\n")
	for _, r := range results {
		if r.Content == "" {
			continue
		}
		fmt.Fprintf(&b, "\n### [%s] %s\n", r.Collection, r.SourceID)
		b.WriteString(r.Content)
		b.WriteString("\n")
	}
	return b.String()
}
