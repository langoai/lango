package learning

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/llm"
	"github.com/langoai/lango/internal/session"
)

const conversationAnalyzerPrompt = `You are a knowledge extraction assistant. Analyze the following conversation and extract structured knowledge.

For each piece of knowledge found, output a JSON object with these fields:
- "type": one of "rule", "definition", "preference", "fact", "pattern", "correction"
- "category": a brief category label (e.g., "go-style", "api-design", "user-preference")
- "content": the extracted knowledge as a clear, reusable statement
- "confidence": one of "low", "medium", "high"
- "temporal": one of "evergreen" (always-true knowledge like "Go uses gofmt") or "current_state" (may change over time like "the team lead is Alice")
- "subject": (optional) entity for graph triple
- "predicate": (optional) relationship for graph triple
- "object": (optional) target entity for graph triple

Output a JSON array of extracted items. If nothing useful is found, output an empty array [].

Focus on:
- Rules and constraints (invariants, coding standards, project rules)
- Definitions and terminology (domain-specific terms, acronyms)
- User preferences and requirements
- Domain knowledge and facts
- Repeated patterns or workflows
- Corrections where the user corrected the agent
- Tool usage patterns`

// ConversationAnalyzer extracts knowledge from conversation turns using LLM analysis.
type ConversationAnalyzer struct {
	generator llm.TextGenerator
	store     *knowledge.Store
	bus       *eventbus.Bus // Optional event bus for publishing triple events.
	logger    *zap.SugaredLogger
}

// NewConversationAnalyzer creates a new conversation analyzer.
func NewConversationAnalyzer(
	generator llm.TextGenerator,
	store *knowledge.Store,
	logger *zap.SugaredLogger,
) *ConversationAnalyzer {
	return &ConversationAnalyzer{
		generator: generator,
		store:     store,
		logger:    logger,
	}
}

// SetEventBus sets the optional event bus for publishing triple events.
func (a *ConversationAnalyzer) SetEventBus(bus *eventbus.Bus) {
	a.bus = bus
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
	saveAnalysisResult(ctx, a.store, a.bus, a.logger, sessionKey, r, saveResultParams{
		KeyPrefix:     "conv",
		TriggerPrefix: "conversation",
		Source:        "conversation_analysis",
	})
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
