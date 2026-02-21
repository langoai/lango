package knowledge

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent/enttest"
	entlearning "github.com/langowarny/lango/internal/ent/learning"
	_ "github.com/mattn/go-sqlite3"
)

func newTestRetriever(t *testing.T) (*ContextRetriever, *Store) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	store := NewStore(client, logger)
	retriever := NewContextRetriever(store, 5, logger)
	return retriever, store
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		give string
		want []string
	}{
		{
			give: "how to handle errors in Go",
			want: []string{"handle", "errors", "go"},
		},
		{
			give: "the a is are was",
			want: nil,
		},
		{
			give: "",
			want: nil,
		},
		{
			give: "hello, world! testing...",
			want: []string{"hello", "world", "testing"},
		},
		{
			give: "a b c x",
			want: nil,
		},
		{
			give: "Deploy PRODUCTION server",
			want: []string{"deploy", "production", "server"},
		},
		{
			give: "(important) [note] {value}",
			want: []string{"important", "note", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := extractKeywords(tt.give)
			if tt.want == nil {
				if got != nil {
					t.Errorf("want nil, got %v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("want %d keywords %v, got %d keywords %v", len(tt.want), tt.want, len(got), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("keyword[%d]: want %q, got %q", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestContextRetriever_Retrieve(t *testing.T) {
	retriever, store := newTestRetriever(t)
	ctx := context.Background()

	// Seed data
	if err := store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
		Key:      "deployment-rule",
		Category: "rule",
		Content:  "Always deploy with rollback plan",
	}); err != nil {
		t.Fatalf("SaveKnowledge: %v", err)
	}

	if err := store.SaveLearning(ctx, "s1", LearningEntry{
		Trigger:      "deployment failure",
		ErrorPattern: "deploy timeout",
		Fix:          "Increase deployment timeout to 5m",
		Category:     entlearning.CategoryTimeout,
	}); err != nil {
		t.Fatalf("SaveLearning: %v", err)
	}

	if err := store.SaveExternalRef(ctx, "deploy-guide", "url", "https://example.com/deploy", "Deployment best practices guide"); err != nil {
		t.Fatalf("SaveExternalRef: %v", err)
	}

	t.Run("multi-layer retrieval", func(t *testing.T) {
		result, err := retriever.Retrieve(ctx, RetrievalRequest{
			Query:      "rollback",
			SessionKey: "s1",
			Layers: []ContextLayer{
				LayerUserKnowledge,
				LayerAgentLearnings,
				LayerSkillPatterns,
				LayerExternalKnowledge,
			},
		})
		if err != nil {
			t.Fatalf("Retrieve: %v", err)
		}
		if result.TotalItems == 0 {
			t.Fatal("expected at least one item across layers")
		}

		// Check knowledge layer
		if items, ok := result.Items[LayerUserKnowledge]; ok {
			found := false
			for _, item := range items {
				if item.Key == "deployment-rule" {
					found = true
				}
			}
			if !found {
				t.Error("expected deployment-rule in user knowledge layer")
			}
		}
	})

	t.Run("single layer", func(t *testing.T) {
		result, err := retriever.Retrieve(ctx, RetrievalRequest{
			Query:  "deploy canary",
			Layers: []ContextLayer{LayerSkillPatterns},
		})
		if err != nil {
			t.Fatalf("Retrieve: %v", err)
		}
		if _, ok := result.Items[LayerUserKnowledge]; ok {
			t.Error("should not have user knowledge when only skill layer requested")
		}
	})
}

func TestContextRetriever_Retrieve_DefaultLayers(t *testing.T) {
	retriever, store := newTestRetriever(t)
	ctx := context.Background()

	if err := store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
		Key:      "default-layer-test",
		Category: "fact",
		Content:  "Testing default layer selection",
	}); err != nil {
		t.Fatalf("SaveKnowledge: %v", err)
	}

	result, err := retriever.Retrieve(ctx, RetrievalRequest{
		Query: "testing default layer",
	})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}

	// With nil Layers, all 4 default layers should be checked
	// We just verify no error and the result is not nil
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Items == nil {
		t.Fatal("expected non-nil Items map")
	}
}

func TestContextRetriever_Retrieve_StopWordsOnly(t *testing.T) {
	retriever, _ := newTestRetriever(t)
	ctx := context.Background()

	result, err := retriever.Retrieve(ctx, RetrievalRequest{
		Query: "the a is",
	})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if result.TotalItems != 0 {
		t.Errorf("want 0 total items for stop-words-only query, got %d", result.TotalItems)
	}
}

// mockToolProvider implements ToolRegistryProvider for testing.
type mockToolProvider struct {
	tools []ToolDescriptor
}

func (m *mockToolProvider) ListTools() []ToolDescriptor {
	return m.tools
}

func (m *mockToolProvider) SearchTools(query string, limit int) []ToolDescriptor {
	queryLower := strings.ToLower(query)
	var result []ToolDescriptor
	for _, t := range m.tools {
		if len(result) >= limit {
			break
		}
		if strings.Contains(strings.ToLower(t.Name), queryLower) ||
			strings.Contains(strings.ToLower(t.Description), queryLower) {
			result = append(result, t)
		}
	}
	return result
}

// mockRuntimeProvider implements RuntimeContextProvider for testing.
type mockRuntimeProvider struct {
	rc RuntimeContext
}

func (m *mockRuntimeProvider) GetRuntimeContext() RuntimeContext {
	return m.rc
}

func TestContextRetriever_RetrieveTools(t *testing.T) {
	retriever, _ := newTestRetriever(t)
	ctx := context.Background()

	tp := &mockToolProvider{
		tools: []ToolDescriptor{
			{Name: "exec_command", Description: "Execute shell commands"},
			{Name: "read_file", Description: "Read file contents"},
			{Name: "web_search", Description: "Search the web"},
		},
	}
	retriever.WithToolRegistry(tp)

	t.Run("keyword search matches tool", func(t *testing.T) {
		result, err := retriever.Retrieve(ctx, RetrievalRequest{
			Query:  "execute shell command",
			Layers: []ContextLayer{LayerToolRegistry},
		})
		if err != nil {
			t.Fatalf("Retrieve: %v", err)
		}
		items, ok := result.Items[LayerToolRegistry]
		if !ok || len(items) == 0 {
			t.Fatal("expected tool registry items")
		}
		if items[0].Key != "exec_command" {
			t.Errorf("want exec_command, got %s", items[0].Key)
		}
	})

	t.Run("nil provider returns empty", func(t *testing.T) {
		r2, _ := newTestRetriever(t)
		result, err := r2.Retrieve(ctx, RetrievalRequest{
			Query:  "execute shell",
			Layers: []ContextLayer{LayerToolRegistry},
		})
		if err != nil {
			t.Fatalf("Retrieve: %v", err)
		}
		if result.TotalItems != 0 {
			t.Errorf("want 0 items with nil provider, got %d", result.TotalItems)
		}
	})
}

func TestContextRetriever_RetrieveRuntimeContext(t *testing.T) {
	retriever, _ := newTestRetriever(t)
	ctx := context.Background()

	rp := &mockRuntimeProvider{
		rc: RuntimeContext{
			SessionKey:        "telegram:123:456",
			ChannelType:       "telegram",
			ActiveToolCount:   5,
			EncryptionEnabled: true,
			KnowledgeEnabled:  true,
			MemoryEnabled:     false,
		},
	}
	retriever.WithRuntimeContext(rp)

	result, err := retriever.Retrieve(ctx, RetrievalRequest{
		Query:  "session state information",
		Layers: []ContextLayer{LayerRuntimeContext},
	})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	items, ok := result.Items[LayerRuntimeContext]
	if !ok || len(items) == 0 {
		t.Fatal("expected runtime context items")
	}
	if !strings.Contains(items[0].Content, "telegram:123:456") {
		t.Errorf("expected session key in content, got %q", items[0].Content)
	}
	if !strings.Contains(items[0].Content, "Channel: telegram") {
		t.Errorf("expected channel type in content, got %q", items[0].Content)
	}

	t.Run("nil provider returns empty", func(t *testing.T) {
		r2, _ := newTestRetriever(t)
		result, err := r2.Retrieve(ctx, RetrievalRequest{
			Query:  "session info",
			Layers: []ContextLayer{LayerRuntimeContext},
		})
		if err != nil {
			t.Fatalf("Retrieve: %v", err)
		}
		if result.TotalItems != 0 {
			t.Errorf("want 0 items with nil provider, got %d", result.TotalItems)
		}
	})
}

func TestAssemblePrompt_WithToolsAndRuntime(t *testing.T) {
	retriever, _ := newTestRetriever(t)
	basePrompt := "You are a helpful assistant."

	result := &RetrievalResult{
		Items: map[ContextLayer][]ContextItem{
			LayerRuntimeContext: {
				{Layer: LayerRuntimeContext, Key: "session-state", Content: "Session: s1 | Channel: telegram | Tools: 3 | Encryption: true | Knowledge: true | Memory: false"},
			},
			LayerToolRegistry: {
				{Layer: LayerToolRegistry, Key: "exec_command", Content: "Execute shell commands"},
			},
			LayerUserKnowledge: {
				{Layer: LayerUserKnowledge, Key: "rule-1", Content: "Always test", Category: "rule"},
			},
		},
		TotalItems: 3,
	}
	got := retriever.AssemblePrompt(basePrompt, result)

	// Runtime Context should appear before User Knowledge
	runtimeIdx := strings.Index(got, "## Runtime Context")
	toolIdx := strings.Index(got, "## Available Tools")
	knowledgeIdx := strings.Index(got, "## User Knowledge")

	if runtimeIdx == -1 {
		t.Error("expected Runtime Context section")
	}
	if toolIdx == -1 {
		t.Error("expected Available Tools section")
	}
	if knowledgeIdx == -1 {
		t.Error("expected User Knowledge section")
	}
	if runtimeIdx > toolIdx {
		t.Error("Runtime Context should appear before Available Tools")
	}
	if toolIdx > knowledgeIdx {
		t.Error("Available Tools should appear before User Knowledge")
	}
	if !strings.Contains(got, "exec_command") {
		t.Error("expected tool name in output")
	}
	if !strings.Contains(got, "Channel: telegram") {
		t.Error("expected runtime info in output")
	}
}

func TestAssemblePrompt(t *testing.T) {
	retriever, _ := newTestRetriever(t)
	basePrompt := "You are a helpful assistant."

	t.Run("nil result returns base", func(t *testing.T) {
		got := retriever.AssemblePrompt(basePrompt, nil)
		if got != basePrompt {
			t.Errorf("want base prompt, got %q", got)
		}
	})

	t.Run("empty items returns base", func(t *testing.T) {
		result := &RetrievalResult{
			Items:      map[ContextLayer][]ContextItem{},
			TotalItems: 0,
		}
		got := retriever.AssemblePrompt(basePrompt, result)
		if got != basePrompt {
			t.Errorf("want base prompt, got %q", got)
		}
	})

	t.Run("4-layer result formats correctly", func(t *testing.T) {
		result := &RetrievalResult{
			Items: map[ContextLayer][]ContextItem{
				LayerUserKnowledge: {
					{Layer: LayerUserKnowledge, Key: "rule-1", Content: "Always test", Category: "rule"},
				},
				LayerAgentLearnings: {
					{Layer: LayerAgentLearnings, Key: "learning-1", Content: "When 'timeout' occurs: increase timeout"},
				},
				LayerSkillPatterns: {
					{Layer: LayerSkillPatterns, Key: "deploy", Content: "Deploy service"},
				},
				LayerExternalKnowledge: {
					{Layer: LayerExternalKnowledge, Key: "docs", Content: "API reference", Source: "https://example.com"},
				},
			},
			TotalItems: 4,
		}
		got := retriever.AssemblePrompt(basePrompt, result)

		if !strings.HasPrefix(got, basePrompt) {
			t.Error("assembled prompt should start with base prompt")
		}
		if !strings.Contains(got, "## User Knowledge") {
			t.Error("expected User Knowledge section")
		}
		if !strings.Contains(got, "## Known Solutions") {
			t.Error("expected Known Solutions section")
		}
		if !strings.Contains(got, "## Available Skills") {
			t.Error("expected Available Skills section")
		}
		if !strings.Contains(got, "## External References") {
			t.Error("expected External References section")
		}
		if !strings.Contains(got, "rule-1") {
			t.Error("expected rule-1 in output")
		}
		if !strings.Contains(got, "deploy") {
			t.Error("expected deploy in output")
		}
		if !strings.Contains(got, "https://example.com") {
			t.Error("expected source URL in output")
		}
	})
}
