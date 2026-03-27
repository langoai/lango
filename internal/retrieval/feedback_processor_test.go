package retrieval

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/langoai/lango/internal/eventbus"
)

func TestFeedbackProcessor_HandleContextInjected(t *testing.T) {
	tests := []struct {
		give     string
		giveEvt  eventbus.ContextInjectedEvent
		wantKeys []string
		wantNoKeys []string
	}{
		{
			give: "event with items and turn ID",
			giveEvt: eventbus.ContextInjectedEvent{
				TurnID:     "turn-123",
				SessionKey: "session-abc",
				Query:      "test query",
				Items: []eventbus.ContextInjectedItem{
					{Layer: "user_knowledge", Key: "k1", Score: 0.9, Source: "fts5", Category: "fact", TokenEstimate: 50},
					{Layer: "agent_learnings", Key: "l1", Score: 0.7, Source: "like", Category: "tool_error", TokenEstimate: 30},
				},
				KnowledgeTokens:  80,
				RAGTokens:        0,
				MemoryTokens:     120,
				RunSummaryTokens: 40,
				TotalTokens:      240,
				Timestamp:        time.Now(),
			},
			wantKeys:   []string{"turn_id", "session_key", "knowledge_items", "total_tokens", "layer_distribution", "source_distribution"},
			wantNoKeys: []string{"query"},
		},
		{
			give: "event with empty items",
			giveEvt: eventbus.ContextInjectedEvent{
				TurnID:     "turn-456",
				SessionKey: "session-def",
				Query:      "",
				Items:      nil,
				Timestamp:  time.Now(),
			},
			wantKeys:   []string{"turn_id", "knowledge_items"},
			wantNoKeys: []string{"query"},
		},
		{
			give: "event without turn ID (direct call)",
			giveEvt: eventbus.ContextInjectedEvent{
				TurnID:     "",
				SessionKey: "session-ghi",
				Query:      "some query",
				Items: []eventbus.ContextInjectedItem{
					{Layer: "user_knowledge", Key: "k2", Score: 0.5, Source: "like"},
				},
				KnowledgeTokens: 50,
				TotalTokens:     50,
				Timestamp:       time.Now(),
			},
			wantKeys:   []string{"session_key", "knowledge_items"},
			wantNoKeys: []string{"turn_id", "query"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			core, recorded := observer.New(zap.InfoLevel)
			logger := zap.New(core).Sugar()

			p := NewFeedbackProcessor(logger)

			bus := eventbus.New()
			p.Subscribe(bus)
			bus.Publish(tt.giveEvt)

			entries := recorded.All()
			assert.NotEmpty(t, entries, "expected log entry")
			assert.Equal(t, "context injected", entries[0].Message)

			contextMap := entries[0].ContextMap()
			for _, key := range tt.wantKeys {
				_, ok := contextMap[key]
				assert.True(t, ok, "expected log key %q", key)
			}
			for _, key := range tt.wantNoKeys {
				_, ok := contextMap[key]
				assert.False(t, ok, "unexpected log key %q (PII or omitted)", key)
			}
		})
	}
}

func TestFeedbackProcessor_NoTurnID_SkipsCorrelation(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core).Sugar()

	p := NewFeedbackProcessor(logger)
	bus := eventbus.New()
	p.Subscribe(bus)

	bus.Publish(eventbus.ContextInjectedEvent{
		TurnID:     "",
		SessionKey: "session-no-turn",
		Timestamp:  time.Now(),
	})

	entries := recorded.All()
	assert.NotEmpty(t, entries)

	contextMap := entries[0].ContextMap()
	_, hasTurnID := contextMap["turn_id"]
	assert.False(t, hasTurnID, "turn_id should be omitted when empty")
}

func TestFeedbackProcessor_LayerDistribution(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core).Sugar()

	p := NewFeedbackProcessor(logger)
	bus := eventbus.New()
	p.Subscribe(bus)

	bus.Publish(eventbus.ContextInjectedEvent{
		TurnID:     "turn-dist",
		SessionKey: "session-dist",
		Items: []eventbus.ContextInjectedItem{
			{Layer: "user_knowledge", Key: "k1", Source: "fts5"},
			{Layer: "user_knowledge", Key: "k2", Source: "fts5"},
			{Layer: "agent_learnings", Key: "l1", Source: "like"},
		},
		Timestamp: time.Now(),
	})

	entries := recorded.All()
	assert.NotEmpty(t, entries)

	contextMap := entries[0].ContextMap()
	assert.Equal(t, int64(3), contextMap["knowledge_items"])
}
