package token

import (
	"testing"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/observability"
)

type mockStore struct {
	saved []observability.TokenUsage
}

func (m *mockStore) Save(usage observability.TokenUsage) error {
	m.saved = append(m.saved, usage)
	return nil
}

func TestTracker_Handle(t *testing.T) {
	tests := []struct {
		give       eventbus.TokenUsageEvent
		wantInput  int64
		wantOutput int64
		wantStored bool
	}{
		{
			give: eventbus.TokenUsageEvent{
				Provider:     "openai",
				Model:        "gpt-4o",
				SessionKey:   "sess-1",
				AgentName:    "main",
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
			},
			wantInput:  100,
			wantOutput: 50,
			wantStored: true,
		},
		{
			give: eventbus.TokenUsageEvent{
				Provider:     "anthropic",
				Model:        "claude-3",
				InputTokens:  200,
				OutputTokens: 100,
				TotalTokens:  300,
				CacheTokens:  50,
			},
			wantInput:  200,
			wantOutput: 100,
			wantStored: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give.Provider+"/"+tt.give.Model, func(t *testing.T) {
			collector := observability.NewCollector()
			store := &mockStore{}

			tracker := NewTracker(collector, store)

			bus := eventbus.New()
			tracker.Subscribe(bus)
			bus.Publish(tt.give)

			snap := collector.Snapshot()
			if snap.TokenUsageTotal.InputTokens != tt.wantInput {
				t.Errorf("InputTokens = %d, want %d", snap.TokenUsageTotal.InputTokens, tt.wantInput)
			}
			if snap.TokenUsageTotal.OutputTokens != tt.wantOutput {
				t.Errorf("OutputTokens = %d, want %d", snap.TokenUsageTotal.OutputTokens, tt.wantOutput)
			}

			if tt.wantStored && len(store.saved) != 1 {
				t.Errorf("store.saved count = %d, want 1", len(store.saved))
			}
		})
	}
}

func TestTracker_NilStore(t *testing.T) {
	collector := observability.NewCollector()
	tracker := NewTracker(collector, nil)

	bus := eventbus.New()
	tracker.Subscribe(bus)
	bus.Publish(eventbus.TokenUsageEvent{
		Provider:     "openai",
		Model:        "gpt-4o",
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	})

	snap := collector.Snapshot()
	if snap.TokenUsageTotal.InputTokens != 100 {
		t.Errorf("InputTokens = %d, want 100", snap.TokenUsageTotal.InputTokens)
	}
}
