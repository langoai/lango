package retrieval

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/knowledge"
)

// mockFactSearchSource is a test double for FactSearchSource.
type mockFactSearchSource struct {
	knowledgeEntries []knowledge.ScoredKnowledgeEntry
	knowledgeErr     error
	learningEntries  []knowledge.ScoredLearningEntry
	learningErr      error
	externalRefs     []knowledge.ExternalRefEntry
	externalRefsErr  error
}

func (m *mockFactSearchSource) SearchKnowledgeScored(_ context.Context, _, _ string, _ int) ([]knowledge.ScoredKnowledgeEntry, error) {
	return m.knowledgeEntries, m.knowledgeErr
}

func (m *mockFactSearchSource) SearchLearningsScored(_ context.Context, _, _ string, _ int) ([]knowledge.ScoredLearningEntry, error) {
	return m.learningEntries, m.learningErr
}

func (m *mockFactSearchSource) SearchExternalRefs(_ context.Context, _ string) ([]knowledge.ExternalRefEntry, error) {
	return m.externalRefs, m.externalRefsErr
}

func TestFactSearchAgent_Name(t *testing.T) {
	agent := NewFactSearchAgent(&mockFactSearchSource{})
	if agent.Name() != "fact-search" {
		t.Errorf("want name %q, got %q", "fact-search", agent.Name())
	}
}

func TestFactSearchAgent_Layers(t *testing.T) {
	agent := NewFactSearchAgent(&mockFactSearchSource{})
	layers := agent.Layers()

	want := []knowledge.ContextLayer{
		knowledge.LayerUserKnowledge,
		knowledge.LayerAgentLearnings,
		knowledge.LayerExternalKnowledge,
	}

	if len(layers) != len(want) {
		t.Fatalf("want %d layers, got %d", len(want), len(layers))
	}
	for i, l := range layers {
		if l != want[i] {
			t.Errorf("layer[%d]: want %v, got %v", i, want[i], l)
		}
	}
}

func TestFactSearchAgent_Search(t *testing.T) {
	tests := []struct {
		give        string
		source      *mockFactSearchSource
		wantCount   int
		wantLayers  map[knowledge.ContextLayer]int
		wantErr     bool
		wantContent map[string]string // key -> expected content
	}{
		{
			give: "all sources return results",
			source: &mockFactSearchSource{
				knowledgeEntries: []knowledge.ScoredKnowledgeEntry{
					{
						Entry:        knowledge.KnowledgeEntry{Key: "k1", Content: "knowledge content", Category: "rule"},
						Score:        0.9,
						SearchSource: "like",
					},
				},
				learningEntries: []knowledge.ScoredLearningEntry{
					{
						Entry:        knowledge.LearningEntry{Trigger: "err pattern", Fix: "apply fix"},
						Score:        0.8,
						SearchSource: "like",
					},
				},
				externalRefs: []knowledge.ExternalRefEntry{
					{Name: "ref1", Summary: "external summary", RefType: "url"},
				},
			},
			wantCount: 3,
			wantLayers: map[knowledge.ContextLayer]int{
				knowledge.LayerUserKnowledge:     1,
				knowledge.LayerAgentLearnings:    1,
				knowledge.LayerExternalKnowledge: 1,
			},
			wantContent: map[string]string{
				"k1":          "knowledge content",
				"err pattern": "When 'err pattern' occurs: apply fix",
				"ref1":        "external summary",
			},
		},
		{
			give: "learning without fix uses trigger as content",
			source: &mockFactSearchSource{
				learningEntries: []knowledge.ScoredLearningEntry{
					{
						Entry:        knowledge.LearningEntry{Trigger: "trigger only"},
						Score:        0.5,
						SearchSource: "like",
					},
				},
			},
			wantCount: 1,
			wantContent: map[string]string{
				"trigger only": "trigger only",
			},
		},
		{
			give: "empty results",
			source: &mockFactSearchSource{},
			wantCount: 0,
		},
		{
			give: "knowledge error propagates",
			source: &mockFactSearchSource{
				knowledgeErr: context.DeadlineExceeded,
			},
			wantErr: true,
		},
		{
			give: "learning error propagates",
			source: &mockFactSearchSource{
				knowledgeEntries: []knowledge.ScoredKnowledgeEntry{},
				learningErr:      context.Canceled,
			},
			wantErr: true,
		},
		{
			give: "external ref error propagates",
			source: &mockFactSearchSource{
				knowledgeEntries: []knowledge.ScoredKnowledgeEntry{},
				learningEntries:  []knowledge.ScoredLearningEntry{},
				externalRefsErr:  context.DeadlineExceeded,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			agent := NewFactSearchAgent(tt.source)
			findings, err := agent.Search(context.Background(), "test query", 10)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(findings) != tt.wantCount {
				t.Fatalf("want %d findings, got %d", tt.wantCount, len(findings))
			}

			// Check layer distribution
			if tt.wantLayers != nil {
				layerCounts := make(map[knowledge.ContextLayer]int)
				for _, f := range findings {
					layerCounts[f.Layer]++
				}
				for layer, wantN := range tt.wantLayers {
					if layerCounts[layer] != wantN {
						t.Errorf("layer %v: want %d findings, got %d", layer, wantN, layerCounts[layer])
					}
				}
			}

			// Check content mapping
			if tt.wantContent != nil {
				contentByKey := make(map[string]string)
				for _, f := range findings {
					contentByKey[f.Key] = f.Content
				}
				for key, wantContent := range tt.wantContent {
					if contentByKey[key] != wantContent {
						t.Errorf("key %q: want content %q, got %q", key, wantContent, contentByKey[key])
					}
				}
			}

			// Verify agent name is set on all findings
			for _, f := range findings {
				if f.Agent != "fact-search" {
					t.Errorf("finding %q: want agent %q, got %q", f.Key, "fact-search", f.Agent)
				}
			}
		})
	}
}

func TestFactSearchAgent_Search_Scores(t *testing.T) {
	source := &mockFactSearchSource{
		knowledgeEntries: []knowledge.ScoredKnowledgeEntry{
			{
				Entry:        knowledge.KnowledgeEntry{Key: "k1", Content: "test"},
				Score:        0.95,
				SearchSource: "like",
			},
		},
		learningEntries: []knowledge.ScoredLearningEntry{
			{
				Entry:        knowledge.LearningEntry{Trigger: "t1"},
				Score:        0.75,
				SearchSource: "like",
			},
		},
		externalRefs: []knowledge.ExternalRefEntry{
			{Name: "r1", Summary: "ref"},
		},
	}

	agent := NewFactSearchAgent(source)
	findings, err := agent.Search(context.Background(), "test", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scores := make(map[string]float64)
	sources := make(map[string]string)
	for _, f := range findings {
		scores[f.Key] = f.Score
		sources[f.Key] = f.SearchSource
	}

	if scores["k1"] != 0.95 {
		t.Errorf("k1 score: want 0.95, got %f", scores["k1"])
	}
	if scores["t1"] != 0.75 {
		t.Errorf("t1 score: want 0.75, got %f", scores["t1"])
	}
	if scores["r1"] != 0 {
		t.Errorf("r1 score: want 0, got %f", scores["r1"])
	}
	if sources["r1"] != "like" {
		t.Errorf("r1 source: want %q, got %q", "like", sources["r1"])
	}
}
