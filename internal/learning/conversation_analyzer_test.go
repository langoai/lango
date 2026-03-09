package learning

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/session"
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
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger)

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
	require.NoError(t, err)

	// Verify knowledge was saved.
	entries, err := store.SearchKnowledge(ctx, "Go modules", "", 10)
	require.NoError(t, err)
	require.NotEmpty(t, entries, "expected at least one knowledge entry after analysis")
}

func TestConversationAnalyzer_Analyze_Correction(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger)

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
	require.NoError(t, err)

	// Verify learning was saved — search by trigger prefix used in saveResult.
	learnings, err := store.SearchLearnings(ctx, "conversation:style", "", 10)
	require.NoError(t, err)
	require.NotEmpty(t, learnings, "expected at least one learning entry after correction analysis")
}

func TestConversationAnalyzer_Analyze_EmptyMessages(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	gen := &fakeTextGenerator{response: "[]"}
	analyzer := NewConversationAnalyzer(gen, nil, logger)

	err := analyzer.Analyze(context.Background(), "test", nil)
	require.NoError(t, err)
}

func TestConversationAnalyzer_Analyze_InvalidJSON(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger)

	gen := &fakeTextGenerator{response: "not valid json at all"}
	analyzer := NewConversationAnalyzer(gen, store, logger)

	msgs := []session.Message{
		{Role: "user", Content: "hello"},
	}

	// Should not error — invalid JSON is non-fatal.
	err := analyzer.Analyze(context.Background(), "test", msgs)
	require.NoError(t, err)
}

func TestConversationAnalyzer_GraphCallback(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger)

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
	require.NoError(t, analyzer.Analyze(ctx, "test", msgs))

	require.NotEmpty(t, callbackTriples, "expected graph callback to receive triples")
	assert.Equal(t, "service:A", callbackTriples[0].Subject)
}
