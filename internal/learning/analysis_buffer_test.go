package learning

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent/enttest"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/session"
	_ "github.com/mattn/go-sqlite3"
)

func TestAnalysisBuffer_StartStop(t *testing.T) {
	logger := zap.NewNop().Sugar()
	gen := &fakeTextGenerator{response: "[]"}

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	store := knowledge.NewStore(client, logger, 20, 10)

	analyzer := NewConversationAnalyzer(gen, store, logger)
	learner := NewSessionLearner(gen, store, logger)

	getMessages := func(_ string) ([]session.Message, error) {
		return nil, nil
	}

	buf := NewAnalysisBuffer(analyzer, learner, getMessages, 10, 2000, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	// Should stop cleanly.
	buf.Stop()
	wg.Wait()
}

func TestAnalysisBuffer_TriggerAnalysis(t *testing.T) {
	logger := zap.NewNop().Sugar()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	store := knowledge.NewStore(client, logger, 20, 10)

	results := []analysisResult{
		{Type: "fact", Category: "test", Content: "buffer analysis fact", Confidence: "high"},
	}
	responseJSON, _ := json.Marshal(results)
	gen := &fakeTextGenerator{response: string(responseJSON)}

	analyzer := NewConversationAnalyzer(gen, store, logger)
	learner := NewSessionLearner(gen, store, logger)

	// Create 15 messages to exceed turn threshold of 10.
	msgs := make([]session.Message, 15)
	for i := range msgs {
		msgs[i] = session.Message{Role: "user", Content: "test message content for analysis"}
	}

	getMessages := func(_ string) ([]session.Message, error) {
		return msgs, nil
	}

	buf := NewAnalysisBuffer(analyzer, learner, getMessages, 10, 2000, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.Trigger("test-session")

	// Wait for processing.
	time.Sleep(200 * time.Millisecond)
	buf.Stop()
	wg.Wait()

	// Verify knowledge was extracted.
	ctx := context.Background()
	entries, err := store.SearchKnowledge(ctx, "buffer analysis", "", 10)
	if err != nil {
		t.Fatalf("SearchKnowledge: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected knowledge entry from buffer analysis trigger")
	}
}

func TestAnalysisBuffer_SessionEnd(t *testing.T) {
	logger := zap.NewNop().Sugar()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	store := knowledge.NewStore(client, logger, 20, 10)

	results := []analysisResult{
		{Type: "preference", Category: "workflow", Content: "session end preference", Confidence: "high"},
	}
	responseJSON, _ := json.Marshal(results)
	gen := &fakeTextGenerator{response: string(responseJSON)}

	analyzer := NewConversationAnalyzer(gen, store, logger)
	learner := NewSessionLearner(gen, store, logger)

	// 5 messages — above 4-turn minimum for session learner.
	msgs := make([]session.Message, 5)
	for i := range msgs {
		msgs[i] = session.Message{Role: "user", Content: "session message"}
	}

	getMessages := func(_ string) ([]session.Message, error) {
		return msgs, nil
	}

	buf := NewAnalysisBuffer(analyzer, learner, getMessages, 10, 2000, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.TriggerSessionEnd("sess-end")

	time.Sleep(200 * time.Millisecond)
	buf.Stop()
	wg.Wait()

	ctx := context.Background()
	entries, err := store.SearchKnowledge(ctx, "session end", "", 10)
	if err != nil {
		t.Fatalf("SearchKnowledge: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected knowledge entry from session-end analysis")
	}
}

func TestAnalysisBuffer_BelowThreshold(t *testing.T) {
	logger := zap.NewNop().Sugar()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	store := knowledge.NewStore(client, logger, 20, 10)

	gen := &fakeTextGenerator{response: "[]"}
	analyzer := NewConversationAnalyzer(gen, store, logger)
	learner := NewSessionLearner(gen, store, logger)

	// 3 messages — below turn threshold of 10 and likely below token threshold.
	msgs := make([]session.Message, 3)
	for i := range msgs {
		msgs[i] = session.Message{Role: "user", Content: "hi"}
	}

	getMessages := func(_ string) ([]session.Message, error) {
		return msgs, nil
	}

	buf := NewAnalysisBuffer(analyzer, learner, getMessages, 10, 2000, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.Trigger("below-threshold")

	time.Sleep(100 * time.Millisecond)
	buf.Stop()
	wg.Wait()

	// No error and no knowledge saved = success (thresholds not met).
}
