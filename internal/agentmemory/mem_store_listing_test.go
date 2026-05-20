package agentmemory

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/ctxkeys"
)

func TestInMemoryStoreListAgentNamesAndListAll(t *testing.T) {
	store := NewInMemoryStore()
	if names, err := store.ListAgentNames(); err != nil || len(names) != 0 {
		t.Fatalf("ListAgentNames(empty) = %v, %v; want empty nil-error result", names, err)
	}
	if entries, err := store.ListAll("missing"); err != nil || entries != nil {
		t.Fatalf("ListAll(missing) = %v, %v; want nil nil-error result", entries, err)
	}

	if err := store.Save(&Entry{
		AgentName:  "planner",
		Key:        "style",
		Kind:       KindPreference,
		Scope:      ScopeInstance,
		Content:    "Use concise summaries",
		Confidence: 0.9,
	}); err != nil {
		t.Fatalf("Save(planner) returned error: %v", err)
	}
	if err := store.Save(&Entry{
		AgentName:  "reviewer",
		Key:        "focus",
		Kind:       KindFact,
		Scope:      ScopeGlobal,
		Content:    "Check concurrency",
		Confidence: 0.8,
	}); err != nil {
		t.Fatalf("Save(reviewer) returned error: %v", err)
	}

	names, err := store.ListAgentNames()
	if err != nil {
		t.Fatalf("ListAgentNames() returned error: %v", err)
	}
	assertContainsAgentName(t, names, "planner")
	assertContainsAgentName(t, names, "reviewer")

	entries, err := store.ListAll("planner")
	if err != nil {
		t.Fatalf("ListAll(planner) returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("ListAll(planner) len = %d, want 1", len(entries))
	}
	if entries[0].Key != "style" || entries[0].Content != "Use concise summaries" {
		t.Fatalf("ListAll(planner)[0] = %#v, want saved style memory", entries[0])
	}

	entries[0].Content = "mutated"
	again, err := store.Get("planner", "style")
	if err != nil {
		t.Fatalf("Get(planner, style) returned error: %v", err)
	}
	if again.Content != "Use concise summaries" {
		t.Fatalf("ListAll() exposed internal entry, got content %q", again.Content)
	}
}

func TestAgentNameOrDefault(t *testing.T) {
	if got := agentNameOrDefault(context.Background()); got != "default" {
		t.Fatalf("agentNameOrDefault(empty) = %q, want default", got)
	}
	ctx := ctxkeys.WithAgentName(context.Background(), "planner")
	if got := agentNameOrDefault(ctx); got != "planner" {
		t.Fatalf("agentNameOrDefault(named) = %q, want planner", got)
	}
}

func assertContainsAgentName(t *testing.T, names []string, want string) {
	t.Helper()
	for _, name := range names {
		if name == want {
			return
		}
	}
	t.Fatalf("agent names %v do not contain %q", names, want)
}
