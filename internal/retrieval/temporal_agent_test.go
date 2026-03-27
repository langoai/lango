package retrieval

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	"github.com/langoai/lango/internal/knowledge"
)

// mockTemporalSource implements TemporalSearchSource for testing.
type mockTemporalSource struct {
	entries []knowledge.KnowledgeEntry
	err     error
}

func (m *mockTemporalSource) SearchRecentKnowledge(_ context.Context, _ string, _ int) ([]knowledge.KnowledgeEntry, error) {
	return m.entries, m.err
}

func TestTemporalSearchAgent_Name(t *testing.T) {
	agent := NewTemporalSearchAgent(&mockTemporalSource{})
	assert.Equal(t, "temporal-search", agent.Name())
}

func TestTemporalSearchAgent_Layers(t *testing.T) {
	agent := NewTemporalSearchAgent(&mockTemporalSource{})
	layers := agent.Layers()
	require.Len(t, layers, 1)
	assert.Equal(t, knowledge.LayerUserKnowledge, layers[0])
}

func TestTemporalSearchAgent_Search(t *testing.T) {
	fixedNow := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		give      string
		entries   []knowledge.KnowledgeEntry
		wantLen   int
		wantScore float64
		wantAgent string
		wantSrc   string
		wantLayer knowledge.ContextLayer
	}{
		{
			give: "recent entry scored high",
			entries: []knowledge.KnowledgeEntry{
				{
					Key:       "user-lang",
					Category:  entknowledge.CategoryFact,
					Content:   "User prefers Go",
					Version:   3,
					UpdatedAt: fixedNow.Add(-1 * time.Hour),
				},
			},
			wantLen:   1,
			wantScore: 1.0 - 1.0/168.0, // ~0.994
			wantAgent: "temporal-search",
			wantSrc:   "temporal",
			wantLayer: knowledge.LayerUserKnowledge,
		},
		{
			give: "old entry scored low",
			entries: []knowledge.KnowledgeEntry{
				{
					Key:       "old-fact",
					Category:  entknowledge.CategoryFact,
					Content:   "Old info",
					Version:   1,
					UpdatedAt: fixedNow.Add(-144 * time.Hour), // 6 days
				},
			},
			wantLen:   1,
			wantScore: 1.0 - 144.0/168.0, // ~0.143
			wantAgent: "temporal-search",
			wantSrc:   "temporal",
			wantLayer: knowledge.LayerUserKnowledge,
		},
		{
			give: "expired entry scored zero",
			entries: []knowledge.KnowledgeEntry{
				{
					Key:       "ancient",
					Category:  entknowledge.CategoryRule,
					Content:   "Very old",
					Version:   1,
					UpdatedAt: fixedNow.Add(-200 * time.Hour), // >1 week
				},
			},
			wantLen:   1,
			wantScore: 0.0,
			wantAgent: "temporal-search",
			wantSrc:   "temporal",
			wantLayer: knowledge.LayerUserKnowledge,
		},
		{
			give:      "empty results",
			entries:   nil,
			wantLen:   0,
			wantAgent: "temporal-search",
			wantSrc:   "temporal",
			wantLayer: knowledge.LayerUserKnowledge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			agent := NewTemporalSearchAgent(&mockTemporalSource{entries: tt.entries})
			agent.now = func() time.Time { return fixedNow }

			findings, err := agent.Search(context.Background(), "test query", 10)
			require.NoError(t, err)
			assert.Len(t, findings, tt.wantLen)

			if tt.wantLen > 0 {
				f := findings[0]
				assert.InDelta(t, tt.wantScore, f.Score, 0.001)
				assert.Equal(t, tt.wantAgent, f.Agent)
				assert.Equal(t, tt.wantSrc, f.SearchSource)
				assert.Equal(t, tt.wantLayer, f.Layer)
			}
		})
	}
}

func TestTemporalSearchAgent_Search_ContentEnriched(t *testing.T) {
	fixedNow := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	agent := NewTemporalSearchAgent(&mockTemporalSource{
		entries: []knowledge.KnowledgeEntry{
			{
				Key:       "user-lang",
				Category:  entknowledge.CategoryFact,
				Content:   "User prefers Go",
				Version:   5,
				UpdatedAt: fixedNow.Add(-3 * time.Hour),
			},
		},
	})
	agent.now = func() time.Time { return fixedNow }

	findings, err := agent.Search(context.Background(), "lang", 10)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "[v5 | updated 3h ago] User prefers Go", findings[0].Content)
}

func TestTemporalSearchAgent_Search_Error(t *testing.T) {
	agent := NewTemporalSearchAgent(&mockTemporalSource{
		err: errors.New("db failure"),
	})

	findings, err := agent.Search(context.Background(), "test", 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "temporal search")
	assert.Nil(t, findings)
}

func TestRecencyScore(t *testing.T) {
	now := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		give      string
		updatedAt time.Time
		wantScore float64
	}{
		{
			give:      "just now",
			updatedAt: now,
			wantScore: 1.0,
		},
		{
			give:      "halfway through window",
			updatedAt: now.Add(-84 * time.Hour),
			wantScore: 0.5,
		},
		{
			give:      "at boundary",
			updatedAt: now.Add(-168 * time.Hour),
			wantScore: 0.0,
		},
		{
			give:      "past boundary",
			updatedAt: now.Add(-200 * time.Hour),
			wantScore: 0.0,
		},
		{
			give:      "1 hour ago",
			updatedAt: now.Add(-1 * time.Hour),
			wantScore: 1.0 - 1.0/168.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			score := recencyScore(tt.updatedAt, now)
			assert.InDelta(t, tt.wantScore, score, 0.001)
		})
	}
}

func TestEnrichTemporalContent(t *testing.T) {
	now := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		give      string
		content   string
		version   int
		updatedAt time.Time
		want      string
	}{
		{
			give:      "just now",
			content:   "Hello",
			version:   1,
			updatedAt: now.Add(-10 * time.Second),
			want:      "[v1 | updated just now] Hello",
		},
		{
			give:      "minutes ago",
			content:   "World",
			version:   2,
			updatedAt: now.Add(-30 * time.Minute),
			want:      "[v2 | updated 30m ago] World",
		},
		{
			give:      "hours ago",
			content:   "Test",
			version:   5,
			updatedAt: now.Add(-3 * time.Hour),
			want:      "[v5 | updated 3h ago] Test",
		},
		{
			give:      "days ago",
			content:   "Old",
			version:   10,
			updatedAt: now.Add(-72 * time.Hour),
			want:      "[v10 | updated 3d ago] Old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := enrichTemporalContent(tt.content, tt.version, tt.updatedAt, now)
			assert.Equal(t, tt.want, result)
		})
	}
}
