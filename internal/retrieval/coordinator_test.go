package retrieval

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/knowledge"
)

// mockAgent implements RetrievalAgent for testing.
type mockAgent struct {
	name     string
	layers   []knowledge.ContextLayer
	findings []Finding
	err      error
}

func (m *mockAgent) Name() string                    { return m.name }
func (m *mockAgent) Layers() []knowledge.ContextLayer { return m.layers }
func (m *mockAgent) Search(_ context.Context, _ string, _ int) ([]Finding, error) {
	return m.findings, m.err
}

func TestRetrievalCoordinator_Retrieve_Parallel(t *testing.T) {
	agent1 := &mockAgent{
		name:   "agent-1",
		layers: []knowledge.ContextLayer{knowledge.LayerUserKnowledge},
		findings: []Finding{
			{Key: "k1", Content: "from agent 1", Score: 0.9, Layer: knowledge.LayerUserKnowledge, Agent: "agent-1"},
		},
	}
	agent2 := &mockAgent{
		name:   "agent-2",
		layers: []knowledge.ContextLayer{knowledge.LayerAgentLearnings},
		findings: []Finding{
			{Key: "l1", Content: "from agent 2", Score: 0.8, Layer: knowledge.LayerAgentLearnings, Agent: "agent-2"},
		},
	}

	logger := zap.NewNop().Sugar()
	coord := NewRetrievalCoordinator([]RetrievalAgent{agent1, agent2}, logger)

	findings, err := coord.Retrieve(context.Background(), "test", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("want 2 findings, got %d", len(findings))
	}

	// First should be higher score
	if findings[0].Score < findings[1].Score {
		t.Errorf("findings not sorted by score desc: %f < %f", findings[0].Score, findings[1].Score)
	}
}

func TestRetrievalCoordinator_Retrieve_Dedup(t *testing.T) {
	tests := []struct {
		give      string
		agents    []RetrievalAgent
		wantCount int
		wantScore float64 // expected score of the deduped item
	}{
		{
			give: "same layer and key from two agents keeps highest score",
			agents: []RetrievalAgent{
				&mockAgent{
					name: "a1",
					findings: []Finding{
						{Key: "shared", Content: "low", Score: 0.3, Layer: knowledge.LayerUserKnowledge, Agent: "a1"},
					},
				},
				&mockAgent{
					name: "a2",
					findings: []Finding{
						{Key: "shared", Content: "high", Score: 0.9, Layer: knowledge.LayerUserKnowledge, Agent: "a2"},
					},
				},
			},
			wantCount: 1,
			wantScore: 0.9,
		},
		{
			give: "different layers same key keeps both",
			agents: []RetrievalAgent{
				&mockAgent{
					name: "a1",
					findings: []Finding{
						{Key: "same-key", Content: "knowledge", Score: 0.5, Layer: knowledge.LayerUserKnowledge, Agent: "a1"},
					},
				},
				&mockAgent{
					name: "a2",
					findings: []Finding{
						{Key: "same-key", Content: "learning", Score: 0.7, Layer: knowledge.LayerAgentLearnings, Agent: "a2"},
					},
				},
			},
			wantCount: 2,
		},
	}

	logger := zap.NewNop().Sugar()
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			coord := NewRetrievalCoordinator(tt.agents, logger)
			findings, err := coord.Retrieve(context.Background(), "test", 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(findings) != tt.wantCount {
				t.Fatalf("want %d findings, got %d", tt.wantCount, len(findings))
			}

			if tt.wantScore > 0 && findings[0].Score != tt.wantScore {
				t.Errorf("want score %f, got %f", tt.wantScore, findings[0].Score)
			}
		})
	}
}

func TestRetrievalCoordinator_Retrieve_AgentError(t *testing.T) {
	healthy := &mockAgent{
		name: "healthy",
		findings: []Finding{
			{Key: "ok", Content: "works", Score: 0.5, Layer: knowledge.LayerUserKnowledge, Agent: "healthy"},
		},
	}
	broken := &mockAgent{
		name: "broken",
		err:  fmt.Errorf("connection lost"),
	}

	logger := zap.NewNop().Sugar()
	coord := NewRetrievalCoordinator([]RetrievalAgent{healthy, broken}, logger)

	findings, err := coord.Retrieve(context.Background(), "test", 0)
	if err != nil {
		t.Fatalf("expected non-fatal error handling, got: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("want 1 finding from healthy agent, got %d", len(findings))
	}
}

func TestTruncateFindings(t *testing.T) {
	tests := []struct {
		give        string
		findings    []Finding
		tokenBudget int
		wantCount   int
	}{
		{
			give: "all fit within budget",
			findings: []Finding{
				{Key: "a", Content: "short", Score: 0.9},
				{Key: "b", Content: "also short", Score: 0.8},
			},
			tokenBudget: 1000,
			wantCount:   2,
		},
		{
			give: "truncates lowest score when over budget",
			findings: []Finding{
				{Key: "a", Content: "first item with some content", Score: 0.9},
				{Key: "b", Content: "second item with some content that is longer to exceed token budget estimate", Score: 0.8},
			},
			tokenBudget: 10,
			wantCount:   1,
		},
		{
			give:        "zero budget returns all",
			findings:    []Finding{{Key: "a", Content: "test", Score: 0.5}},
			tokenBudget: 0,
			wantCount:   1,
		},
		{
			give:        "empty findings",
			findings:    nil,
			tokenBudget: 100,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := TruncateFindings(tt.findings, tt.tokenBudget)
			if len(result) != tt.wantCount {
				t.Errorf("want %d findings, got %d", tt.wantCount, len(result))
			}

			// Verify highest-score findings are kept
			if len(result) > 0 && len(tt.findings) > 0 {
				if result[0].Key != tt.findings[0].Key {
					t.Errorf("expected highest-score finding %q to be kept, got %q", tt.findings[0].Key, result[0].Key)
				}
			}
		})
	}
}

func TestToRetrievalResult(t *testing.T) {
	findings := []Finding{
		{Key: "k1", Content: "knowledge", Score: 0.9, Category: "rule", SearchSource: "like", Layer: knowledge.LayerUserKnowledge},
		{Key: "k2", Content: "more knowledge", Score: 0.7, Category: "fact", SearchSource: "like", Layer: knowledge.LayerUserKnowledge},
		{Key: "l1", Content: "learning", Score: 0.8, Category: "tool_error", SearchSource: "like", Layer: knowledge.LayerAgentLearnings},
	}

	result := ToRetrievalResult(findings)

	if result.TotalItems != 3 {
		t.Errorf("TotalItems: want 3, got %d", result.TotalItems)
	}

	knowledgeItems := result.Items[knowledge.LayerUserKnowledge]
	if len(knowledgeItems) != 2 {
		t.Fatalf("LayerUserKnowledge: want 2 items, got %d", len(knowledgeItems))
	}

	learningItems := result.Items[knowledge.LayerAgentLearnings]
	if len(learningItems) != 1 {
		t.Fatalf("LayerAgentLearnings: want 1 item, got %d", len(learningItems))
	}

	// Verify Score is propagated to ContextItem
	if knowledgeItems[0].Score != 0.9 {
		t.Errorf("ContextItem.Score: want 0.9, got %f", knowledgeItems[0].Score)
	}
	if knowledgeItems[0].Source != "like" {
		t.Errorf("ContextItem.Source: want %q, got %q", "like", knowledgeItems[0].Source)
	}
	if knowledgeItems[0].Category != "rule" {
		t.Errorf("ContextItem.Category: want %q, got %q", "rule", knowledgeItems[0].Category)
	}
	if learningItems[0].Score != 0.8 {
		t.Errorf("learning ContextItem.Score: want 0.8, got %f", learningItems[0].Score)
	}
}

func TestToRetrievalResult_Empty(t *testing.T) {
	result := ToRetrievalResult(nil)
	if result.TotalItems != 0 {
		t.Errorf("TotalItems: want 0, got %d", result.TotalItems)
	}
	if len(result.Items) != 0 {
		t.Errorf("Items: want empty, got %d layers", len(result.Items))
	}
}

func TestRetrievalCoordinator_Shadow(t *testing.T) {
	logger := zap.NewNop().Sugar()
	coord := NewRetrievalCoordinator(nil, logger)

	if !coord.Shadow() {
		t.Error("expected default shadow=true")
	}

	coord.SetShadow(false)
	if coord.Shadow() {
		t.Error("expected shadow=false after SetShadow(false)")
	}
}
