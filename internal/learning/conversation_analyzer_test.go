package learning

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent/enttest"
	"github.com/langowarny/lango/internal/graph"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/session"
	_ "github.com/mattn/go-sqlite3"
)

// fakeTextGenerator returns a predefined response.
type fakeTextGenerator struct {
	response string
	err      error
}

func (g *fakeTextGenerator) GenerateText(_ context.Context, _, _ string) (string, error) {
	return g.response, g.err
}

func TestConversationAnalyzer_Analyze_Fact(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger, 20, 10)

	results := []analysisResult{
		{Type: "fact", Category: "domain", Content: "User prefers Go modules", Confidence: "high"},
	}
	responseJSON, _ := json.Marshal(results)

	gen := &fakeTextGenerator{response: string(responseJSON)}
	analyzer := NewConversationAnalyzer(gen, store, logger)

	msgs := []session.Message{
		{Role: "user", Content: "I always use Go modules for dependency management"},
		{Role: "assistant", Content: "Understood, I will use Go modules."},
	}

	ctx := context.Background()
	err := analyzer.Analyze(ctx, "test-session", msgs)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// Verify knowledge was saved.
	entries, err := store.SearchKnowledge(ctx, "Go modules", "", 10)
	if err != nil {
		t.Fatalf("SearchKnowledge: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one knowledge entry after analysis")
	}
}

func TestConversationAnalyzer_Analyze_Correction(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger, 20, 10)

	results := []analysisResult{
		{Type: "correction", Category: "style", Content: "Use snake_case not camelCase", Confidence: "high"},
	}
	responseJSON, _ := json.Marshal(results)

	gen := &fakeTextGenerator{response: string(responseJSON)}
	analyzer := NewConversationAnalyzer(gen, store, logger)

	msgs := []session.Message{
		{Role: "user", Content: "No, use snake_case not camelCase"},
	}

	ctx := context.Background()
	err := analyzer.Analyze(ctx, "test-session", msgs)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// Verify learning was saved — search by trigger prefix used in saveResult.
	learnings, err := store.SearchLearnings(ctx, "conversation:style", "", 10)
	if err != nil {
		t.Fatalf("SearchLearnings: %v", err)
	}
	if len(learnings) == 0 {
		t.Fatal("expected at least one learning entry after correction analysis")
	}
}

func TestConversationAnalyzer_Analyze_EmptyMessages(t *testing.T) {
	logger := zap.NewNop().Sugar()
	gen := &fakeTextGenerator{response: "[]"}
	analyzer := NewConversationAnalyzer(gen, nil, logger)

	err := analyzer.Analyze(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("Analyze with empty messages should not error: %v", err)
	}
}

func TestConversationAnalyzer_Analyze_InvalidJSON(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger, 20, 10)

	gen := &fakeTextGenerator{response: "not valid json at all"}
	analyzer := NewConversationAnalyzer(gen, store, logger)

	msgs := []session.Message{
		{Role: "user", Content: "hello"},
	}

	// Should not error — invalid JSON is non-fatal.
	err := analyzer.Analyze(context.Background(), "test", msgs)
	if err != nil {
		t.Fatalf("Analyze should not error on invalid JSON: %v", err)
	}
}

func TestConversationAnalyzer_GraphCallback(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger, 20, 10)

	results := []analysisResult{
		{
			Type: "fact", Category: "arch", Content: "Service A depends on Service B",
			Subject: "service:A", Predicate: "depends_on", Object: "service:B",
		},
	}
	responseJSON, _ := json.Marshal(results)
	gen := &fakeTextGenerator{response: string(responseJSON)}

	var callbackTriples []graph.Triple
	analyzer := NewConversationAnalyzer(gen, store, logger)
	analyzer.SetGraphCallback(func(triples []graph.Triple) {
		callbackTriples = append(callbackTriples, triples...)
	})

	msgs := []session.Message{
		{Role: "user", Content: "Service A depends on Service B"},
	}

	ctx := context.Background()
	if err := analyzer.Analyze(ctx, "test", msgs); err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if len(callbackTriples) == 0 {
		t.Fatal("expected graph callback to receive triples")
	}
	if callbackTriples[0].Subject != "service:A" {
		t.Errorf("want subject %q, got %q", "service:A", callbackTriples[0].Subject)
	}
}
