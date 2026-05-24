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

// retrieveGraphRAGData fetches graph-enhanced retrieved context results.
func (m *ContextAwareModelAdapter) retrieveGraphRAGData(ctx context.Context, query, sessionKey string) *graph.GraphRAGResult {
	opts := graph.ContentRetrieveOptions{
		Limit:      m.retrievedLimit,
		SessionKey: sessionKey,
	}
	result, err := m.graphRAG.Retrieve(ctx, query, opts)
	if err != nil {
		m.logger.Warnw("graph rag retrieval error", "error", err)
		return nil
	}
	return result
}

// formatGraphRAGSection truncates and formats a GraphRAG result into a prompt section.
// budgetTokens controls item-level truncation: 0 = unlimited.
func (m *ContextAwareModelAdapter) formatGraphRAGSection(result *graph.GraphRAGResult, budgetTokens int) string {
	if result == nil {
		return ""
	}

	if budgetTokens > 0 {
		remaining := budgetTokens
		for i, r := range result.ContentResults {
			itemTokens := types.EstimateTokens(r.Content) + types.EstimateTokens(fmt.Sprintf("\n### [%s] %s\n", r.Collection, r.SourceID))
			if remaining-itemTokens < 0 {
				result.ContentResults = result.ContentResults[:i]
				result.GraphResults = nil
				break
			}
			remaining -= itemTokens
		}
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

// formatRecallSection formats prior-session recall matches under the shared
// RAG section budget. Higher-ranked matches are kept first; lower-ranked
// matches drop on overflow. Returns an empty string when nothing fits or
// the input is empty.
func formatRecallSection(matches []RecallMatch, budgetTokens int) string {
	if len(matches) == 0 {
		return ""
	}
	const header = "## Prior Session Recall\n"
	if budgetTokens > 0 {
		remaining := budgetTokens - types.EstimateTokens(header)
		kept := 0
		for i, m := range matches {
			entry := fmt.Sprintf("\n### [%s] (rank=%.2f)\n%s\n", m.SessionKey, m.Rank, m.Summary)
			itemTokens := types.EstimateTokens(entry)
			if remaining-itemTokens < 0 {
				matches = matches[:i]
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
	b.WriteString(header)
	for _, m := range matches {
		if m.Summary == "" {
			continue
		}
		fmt.Fprintf(&b, "\n### [%s] (rank=%.2f)\n", m.SessionKey, m.Rank)
		b.WriteString(m.Summary)
		b.WriteString("\n")
	}
	return b.String()
}
