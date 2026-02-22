package learning

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/session"
	_ "github.com/mattn/go-sqlite3"
)

func TestSessionLearner_LearnFromSession_HighConfidence(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger)

	results := []analysisResult{
		{Type: "preference", Category: "tools", Content: "User prefers vim keybindings", Confidence: "high"},
		{Type: "fact", Category: "domain", Content: "Low confidence fact", Confidence: "low"},
	}
	responseJSON, _ := json.Marshal(results)
	gen := &fakeTextGenerator{response: string(responseJSON)}

	learner := NewSessionLearner(gen, store, logger)

	// 5 messages — above the 4-turn minimum.
	msgs := make([]session.Message, 5)
	for i := range msgs {
		msgs[i] = session.Message{Role: "user", Content: "message content"}
	}

	ctx := context.Background()
	err := learner.LearnFromSession(ctx, "sess-1", msgs)
	if err != nil {
		t.Fatalf("LearnFromSession: %v", err)
	}

	// Only high-confidence results should be stored.
	entries, err := store.SearchKnowledge(ctx, "vim", "", 10)
	if err != nil {
		t.Fatalf("SearchKnowledge: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected high-confidence entry to be stored")
	}

	// Low-confidence should NOT be stored.
	lowEntries, err := store.SearchKnowledge(ctx, "Low confidence fact", "", 10)
	if err != nil {
		t.Fatalf("SearchKnowledge: %v", err)
	}
	if len(lowEntries) > 0 {
		t.Error("low-confidence entry should not be stored")
	}
}

func TestSessionLearner_SkipShortSession(t *testing.T) {
	logger := zap.NewNop().Sugar()
	gen := &fakeTextGenerator{response: "[]"}
	learner := NewSessionLearner(gen, nil, logger)

	// 3 messages — below 4-turn minimum.
	msgs := make([]session.Message, 3)
	for i := range msgs {
		msgs[i] = session.Message{Role: "user", Content: "short"}
	}

	err := learner.LearnFromSession(context.Background(), "short-sess", msgs)
	if err != nil {
		t.Fatalf("LearnFromSession should not error for short sessions: %v", err)
	}
}

func TestSampleMessages(t *testing.T) {
	t.Run("short session returns all", func(t *testing.T) {
		msgs := make([]session.Message, 10)
		for i := range msgs {
			msgs[i] = session.Message{Content: "msg"}
		}
		sampled := sampleMessages(msgs)
		if len(sampled) != 10 {
			t.Errorf("want 10 messages for short session, got %d", len(sampled))
		}
	})

	t.Run("exactly 20 returns all", func(t *testing.T) {
		msgs := make([]session.Message, 20)
		for i := range msgs {
			msgs[i] = session.Message{Content: "msg"}
		}
		sampled := sampleMessages(msgs)
		if len(sampled) != 20 {
			t.Errorf("want 20 messages for 20-message session, got %d", len(sampled))
		}
	})

	t.Run("long session samples", func(t *testing.T) {
		msgs := make([]session.Message, 50)
		for i := range msgs {
			msgs[i] = session.Message{Content: "msg"}
		}
		sampled := sampleMessages(msgs)

		// Should be less than total.
		if len(sampled) >= 50 {
			t.Errorf("sampled %d messages from 50, expected fewer", len(sampled))
		}

		// First 3 should be first 3 messages.
		for i := 0; i < 3; i++ {
			if &sampled[i] == &msgs[i] {
				continue // same reference = good
			}
		}

		// Last 5 should be last 5 messages.
		for i := 0; i < 5; i++ {
			sampledIdx := len(sampled) - 5 + i
			msgsIdx := len(msgs) - 5 + i
			if sampled[sampledIdx].Content != msgs[msgsIdx].Content {
				t.Errorf("last-5 mismatch at position %d", i)
			}
		}
	})
}
