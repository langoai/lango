package adk

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/graph"
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
func (m *ContextAwareModelAdapter) assembleGraphRAGSection(ctx context.Context, query, sessionKey string) string {
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
	return m.graphRAG.AssembleSection(result)
}

// assembleRAGSection builds a "Semantic Context" section from RAG retrieval results.
func (m *ContextAwareModelAdapter) assembleRAGSection(ctx context.Context, query, sessionKey string) string {
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
