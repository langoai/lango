package learning

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	entlearning "github.com/langowarny/lango/internal/ent/learning"
	"github.com/langowarny/lango/internal/graph"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/session"
)

// TextGenerator generates text from a prompt (mirrors memory.TextGenerator to avoid import cycle).
type TextGenerator interface {
	GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

const conversationAnalyzerPrompt = `You are a knowledge extraction assistant. Analyze the following conversation and extract structured knowledge.

For each piece of knowledge found, output a JSON object with these fields:
- "type": one of "fact", "pattern", "correction", "preference"
- "category": a brief category label (e.g., "go-style", "api-design", "user-preference")
- "content": the extracted knowledge as a clear, reusable statement
- "confidence": one of "low", "medium", "high"
- "subject": (optional) entity for graph triple
- "predicate": (optional) relationship for graph triple
- "object": (optional) target entity for graph triple

Output a JSON array of extracted items. If nothing useful is found, output an empty array [].

Focus on:
- User preferences and requirements
- Domain knowledge and facts
- Repeated patterns or workflows
- Corrections where the user corrected the agent
- Tool usage patterns`

// ConversationAnalyzer extracts knowledge from conversation turns using LLM analysis.
type ConversationAnalyzer struct {
	generator     TextGenerator
	store         *knowledge.Store
	graphCallback GraphCallback
	logger        *zap.SugaredLogger
}

// NewConversationAnalyzer creates a new conversation analyzer.
func NewConversationAnalyzer(
	generator TextGenerator,
	store *knowledge.Store,
	logger *zap.SugaredLogger,
) *ConversationAnalyzer {
	return &ConversationAnalyzer{
		generator: generator,
		store:     store,
		logger:    logger,
	}
}

// SetGraphCallback sets the optional graph update hook.
func (a *ConversationAnalyzer) SetGraphCallback(cb GraphCallback) {
	a.graphCallback = cb
}

// Analyze processes a batch of messages and extracts knowledge.
func (a *ConversationAnalyzer) Analyze(ctx context.Context, sessionKey string, messages []session.Message) error {
	if len(messages) == 0 {
		return nil
	}

	userPrompt := formatMessagesForAnalysis(messages)
	response, err := a.generator.GenerateText(ctx, conversationAnalyzerPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("analyze conversation: %w", err)
	}

	results, err := parseAnalysisResponse(response)
	if err != nil {
		a.logger.Debugw("parse analysis response", "error", err, "raw", response)
		return nil // non-fatal — LLM may produce invalid JSON
	}

	for _, r := range results {
		if r.Content == "" {
			continue
		}
		a.saveResult(ctx, sessionKey, r)
	}

	return nil
}

func (a *ConversationAnalyzer) saveResult(ctx context.Context, sessionKey string, r analysisResult) {
	switch r.Type {
	case "fact", "preference":
		key := fmt.Sprintf("conv:%s:%s", sessionKey, sanitizeForNode(r.Content[:min(len(r.Content), 32)]))
		entry := knowledge.KnowledgeEntry{
			Key:      key,
			Category: mapKnowledgeCategory(r.Type),
			Content:  r.Content,
			Source:   "conversation_analysis",
		}
		if err := a.store.SaveKnowledge(ctx, sessionKey, entry); err != nil {
			a.logger.Debugw("save knowledge from analysis", "error", err)
		}

	case "pattern", "correction":
		entry := knowledge.LearningEntry{
			Trigger:   fmt.Sprintf("conversation:%s", r.Category),
			Diagnosis: r.Content,
			Category:  mapLearningCategory(r.Type),
		}
		if r.Type == "correction" {
			entry.Fix = r.Content
			entry.Category = entlearning.CategoryUserCorrection
		}
		if err := a.store.SaveLearning(ctx, sessionKey, entry); err != nil {
			a.logger.Debugw("save learning from analysis", "error", err)
		}
	}

	// Emit graph triples if provided.
	if a.graphCallback != nil && r.Subject != "" && r.Predicate != "" && r.Object != "" {
		a.graphCallback([]graph.Triple{{
			Subject:   r.Subject,
			Predicate: r.Predicate,
			Object:    r.Object,
		}})
	}
}

func formatMessagesForAnalysis(msgs []session.Message) string {
	var b strings.Builder
	for _, msg := range msgs {
		fmt.Fprintf(&b, "[%s]: %s\n", msg.Role, msg.Content)
		for _, tc := range msg.ToolCalls {
			fmt.Fprintf(&b, "  [tool:%s] %s\n", tc.Name, tc.Input)
			if tc.Output != "" {
				out := tc.Output
				if len(out) > 500 {
					out = out[:500] + "..."
				}
				fmt.Fprintf(&b, "  [result] %s\n", out)
			}
		}
	}
	return b.String()
}
