package retrieval

import (
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

// FeedbackProcessor subscribes to ContextInjectedEvent and logs structured
// observability data about which context items were injected into each turn.
//
// This processor is read-only — it does not modify relevance scores or stored data.
// Score auto-adjustment is handled by RelevanceAdjuster (separate subscriber).
type FeedbackProcessor struct {
	logger *zap.SugaredLogger
}

// NewFeedbackProcessor creates a feedback processor for context injection observability.
func NewFeedbackProcessor(logger *zap.SugaredLogger) *FeedbackProcessor {
	return &FeedbackProcessor{logger: logger}
}

// Subscribe registers the processor to receive ContextInjectedEvent from the bus.
func (p *FeedbackProcessor) Subscribe(bus *eventbus.Bus) {
	eventbus.SubscribeTyped[eventbus.ContextInjectedEvent](bus, p.handleContextInjected)
}

func (p *FeedbackProcessor) handleContextInjected(evt eventbus.ContextInjectedEvent) {
	layerCounts := make(map[string]int, len(evt.Items))
	sourceCounts := make(map[string]int, 2)
	for _, item := range evt.Items {
		layerCounts[item.Layer]++
		if item.Source != "" {
			sourceCounts[item.Source]++
		}
	}

	// Note: raw query is NOT logged (PII). Only query_length is recorded.
	// Pre-allocate to avoid slice copy when prepending turn_id.
	fields := make([]interface{}, 0, 22)

	// Include turn_id only when present (TurnRunner sets it; direct calls may not).
	if evt.TurnID != "" {
		fields = append(fields, "turn_id", evt.TurnID)
	}

	fields = append(fields,
		"session_key", evt.SessionKey,
		"query_length", len(evt.Query),
		"knowledge_items", len(evt.Items),
		"knowledge_tokens", evt.KnowledgeTokens,
		"retrieved_tokens", evt.RetrievedTokens,
		"memory_tokens", evt.MemoryTokens,
		"run_summary_tokens", evt.RunSummaryTokens,
		"total_tokens", evt.TotalTokens,
		"layer_distribution", layerCounts,
		"source_distribution", sourceCounts,
	)

	p.logger.Infow("context injected", fields...)
}
