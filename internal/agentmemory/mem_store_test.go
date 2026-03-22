package agentmemory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStore_Save(t *testing.T) {
	tests := []struct {
		give      *Entry
		wantErr   bool
		wantErrAs string
	}{
		{
			give: &Entry{
				AgentName:  "researcher",
				Key:        "go-patterns",
				Content:    "Use table-driven tests",
				Kind:       KindPattern,
				Scope:      ScopeInstance,
				Confidence: 0.9,
			},
		},
		{
			give:      &Entry{Key: "no-agent", Content: "missing agent"},
			wantErr:   true,
			wantErrAs: "agent_name is required",
		},
		{
			give:      &Entry{AgentName: "a", Content: "missing key"},
			wantErr:   true,
			wantErrAs: "key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give.Key, func(t *testing.T) {
			s := NewInMemoryStore()
			err := s.Save(tt.give)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrAs)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestInMemoryStore_Save_Upsert(t *testing.T) {
	s := NewInMemoryStore()

	first := &Entry{
		AgentName:  "researcher",
		Key:        "greeting",
		Content:    "hello",
		Kind:       KindFact,
		Scope:      ScopeInstance,
		Confidence: 0.5,
	}
	require.NoError(t, s.Save(first))

	got, err := s.Get("researcher", "greeting")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "hello", got.Content)
	createdAt := got.CreatedAt

	// Upsert with new content.
	second := &Entry{
		AgentName:  "researcher",
		Key:        "greeting",
		Content:    "updated",
		Kind:       KindPreference,
		Scope:      ScopeGlobal,
		Confidence: 0.8,
	}
	require.NoError(t, s.Save(second))

	got, err = s.Get("researcher", "greeting")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "updated", got.Content)
	assert.Equal(t, KindPreference, got.Kind)
	assert.Equal(t, ScopeGlobal, got.Scope)
	assert.Equal(t, 0.8, got.Confidence)
	assert.Equal(t, createdAt, got.CreatedAt, "CreatedAt should be preserved on upsert")
	assert.True(t, got.UpdatedAt.After(createdAt) || got.UpdatedAt.Equal(createdAt))
}

func TestInMemoryStore_Get(t *testing.T) {
	s := NewInMemoryStore()

	// Get from empty store.
	got, err := s.Get("none", "key")
	require.NoError(t, err)
	assert.Nil(t, got)

	// Save and retrieve.
	require.NoError(t, s.Save(&Entry{
		AgentName: "agent1",
		Key:       "k1",
		Content:   "value1",
		Kind:      KindFact,
		Scope:     ScopeInstance,
	}))

	got, err = s.Get("agent1", "k1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "value1", got.Content)
	assert.NotEmpty(t, got.ID)

	// Get returns a clone (mutations don't affect store).
	got.Content = "mutated"
	original, _ := s.Get("agent1", "k1")
	assert.Equal(t, "value1", original.Content)
}

func TestInMemoryStore_Search(t *testing.T) {
	s := NewInMemoryStore()

	entries := []*Entry{
		{AgentName: "a1", Key: "go-patterns", Content: "table tests", Kind: KindPattern, Scope: ScopeInstance, Confidence: 0.9, Tags: []string{"go", "testing"}},
		{AgentName: "a1", Key: "user-pref", Content: "dark mode", Kind: KindPreference, Scope: ScopeInstance, Confidence: 0.7},
		{AgentName: "a1", Key: "low-conf", Content: "maybe true", Kind: KindFact, Scope: ScopeInstance, Confidence: 0.2},
	}
	for _, e := range entries {
		require.NoError(t, s.Save(e))
	}

	tests := []struct {
		give      SearchOptions
		wantCount int
	}{
		{
			give:      SearchOptions{},
			wantCount: 3,
		},
		{
			give:      SearchOptions{Kind: KindPattern},
			wantCount: 1,
		},
		{
			give:      SearchOptions{MinConfidence: 0.5},
			wantCount: 2,
		},
		{
			give:      SearchOptions{Tags: []string{"go"}},
			wantCount: 1,
		},
		{
			give:      SearchOptions{Query: "dark"},
			wantCount: 1,
		},
		{
			give:      SearchOptions{Limit: 1},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			results, err := s.Search("a1", tt.give)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
		})
	}
}

func TestInMemoryStore_SearchWithContext(t *testing.T) {
	s := NewInMemoryStore()

	// Instance-scoped entry for agent1.
	require.NoError(t, s.Save(&Entry{
		AgentName: "agent1",
		Key:       "local-fact",
		Content:   "instance only",
		Kind:      KindFact,
		Scope:     ScopeInstance,
	}))

	// Global entry from agent2.
	require.NoError(t, s.Save(&Entry{
		AgentName: "agent2",
		Key:       "shared-fact",
		Content:   "global knowledge instance",
		Kind:      KindFact,
		Scope:     ScopeGlobal,
	}))

	// Instance-scoped entry from agent2 (should NOT appear for agent1).
	require.NoError(t, s.Save(&Entry{
		AgentName: "agent2",
		Key:       "private-fact",
		Content:   "agent2 private instance",
		Kind:      KindFact,
		Scope:     ScopeInstance,
	}))

	results, err := s.SearchWithContext("agent1", "instance", 10)
	require.NoError(t, err)

	// Should find: "local-fact" (instance, agent1) + "shared-fact" (global, agent2).
	// Should NOT find: "private-fact" (instance, agent2).
	assert.Len(t, results, 2)

	var keys []string
	for _, r := range results {
		keys = append(keys, r.Key)
	}
	assert.Contains(t, keys, "local-fact")
	assert.Contains(t, keys, "shared-fact")
}

func TestInMemoryStore_Delete(t *testing.T) {
	s := NewInMemoryStore()

	require.NoError(t, s.Save(&Entry{
		AgentName: "a1",
		Key:       "to-delete",
		Content:   "bye",
		Kind:      KindFact,
		Scope:     ScopeInstance,
	}))

	got, err := s.Get("a1", "to-delete")
	require.NoError(t, err)
	require.NotNil(t, got)

	require.NoError(t, s.Delete("a1", "to-delete"))

	got, err = s.Get("a1", "to-delete")
	require.NoError(t, err)
	assert.Nil(t, got)

	// Deleting non-existent entry should not error.
	require.NoError(t, s.Delete("a1", "nonexistent"))
	require.NoError(t, s.Delete("unknown-agent", "any"))
}

func TestInMemoryStore_IncrementUseCount(t *testing.T) {
	s := NewInMemoryStore()

	require.NoError(t, s.Save(&Entry{
		AgentName: "a1",
		Key:       "counter",
		Content:   "test",
		Kind:      KindSkill,
		Scope:     ScopeInstance,
	}))

	require.NoError(t, s.IncrementUseCount("a1", "counter"))
	require.NoError(t, s.IncrementUseCount("a1", "counter"))
	require.NoError(t, s.IncrementUseCount("a1", "counter"))

	got, err := s.Get("a1", "counter")
	require.NoError(t, err)
	assert.Equal(t, 3, got.UseCount)

	// Increment non-existent entry should not error.
	require.NoError(t, s.IncrementUseCount("a1", "nonexistent"))
	require.NoError(t, s.IncrementUseCount("unknown", "any"))
}

func TestInMemoryStore_Prune(t *testing.T) {
	s := NewInMemoryStore()

	entries := []*Entry{
		{AgentName: "a1", Key: "high", Content: "keep", Kind: KindFact, Scope: ScopeInstance, Confidence: 0.9},
		{AgentName: "a1", Key: "mid", Content: "keep", Kind: KindFact, Scope: ScopeInstance, Confidence: 0.5},
		{AgentName: "a1", Key: "low", Content: "prune", Kind: KindFact, Scope: ScopeInstance, Confidence: 0.1},
	}
	for _, e := range entries {
		require.NoError(t, s.Save(e))
	}

	pruned, err := s.Prune("a1", 0.5)
	require.NoError(t, err)
	assert.Equal(t, 1, pruned) // only "low" (0.1) < 0.5

	// Verify remaining entries.
	got, _ := s.Get("a1", "high")
	assert.NotNil(t, got)
	got, _ = s.Get("a1", "mid")
	assert.NotNil(t, got)
	got, _ = s.Get("a1", "low")
	assert.Nil(t, got)

	// Prune non-existent agent should not error.
	pruned, err = s.Prune("unknown", 0.5)
	require.NoError(t, err)
	assert.Equal(t, 0, pruned)
}

// ── Kind validation tests ──

func TestInMemoryStore_Save_InvalidKind(t *testing.T) {
	s := NewInMemoryStore()
	err := s.Save(&Entry{
		AgentName: "agent-1",
		Key:       "test",
		Content:   "content",
		Kind:      MemoryKind("bogus"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid memory kind")
}

func TestInMemoryStore_Save_EmptyKindAllowed(t *testing.T) {
	s := NewInMemoryStore()
	// Empty kind is allowed (defaults happen at handler level).
	err := s.Save(&Entry{
		AgentName: "agent-1",
		Key:       "test",
		Content:   "content",
		Kind:      "",
	})
	require.NoError(t, err)
}

// ── SearchWithContextOptions: kind filter + scope fallback tests ──

func TestSearchWithContextOptions_KindFilterWithGlobalFallback(t *testing.T) {
	s := NewInMemoryStore()

	// Instance entries for agent-1 (different kinds).
	_ = s.Save(&Entry{AgentName: "agent-1", Key: "local-fact", Content: "go is typed", Kind: KindFact, Scope: ScopeInstance})
	_ = s.Save(&Entry{AgentName: "agent-1", Key: "local-pattern", Content: "table driven tests", Kind: KindPattern, Scope: ScopeInstance})

	// Global entry from another agent with matching kind.
	_ = s.Save(&Entry{AgentName: "shared", Key: "global-fact", Content: "go is compiled", Kind: KindFact, Scope: ScopeGlobal})

	results, err := s.SearchWithContextOptions("agent-1", SearchOptions{
		Query: "go",
		Kind:  KindFact,
		Limit: 10,
	})
	require.NoError(t, err)

	// Should include both instance fact AND global fact, but NOT the pattern.
	assert.Len(t, results, 2, "expected instance fact + global fact")
	names := make(map[string]bool)
	for _, r := range results {
		names[r.Key] = true
		assert.Equal(t, KindFact, r.Kind)
	}
	assert.True(t, names["local-fact"], "instance fact should be included")
	assert.True(t, names["global-fact"], "global fact should be included")
}

func TestSearchWithContextOptions_KindFilterSmallLimitWithManyOtherKinds(t *testing.T) {
	s := NewInMemoryStore()

	// Fill instance with many pattern entries to crowd out facts.
	for i := 0; i < 20; i++ {
		_ = s.Save(&Entry{
			AgentName: "agent-1",
			Key:       fmt.Sprintf("pattern-%d", i),
			Content:   "some pattern about go",
			Kind:      KindPattern,
			Scope:     ScopeInstance,
		})
	}

	// One instance fact.
	_ = s.Save(&Entry{AgentName: "agent-1", Key: "instance-fact", Content: "go is fast", Kind: KindFact, Scope: ScopeInstance})

	// One global fact from another agent.
	_ = s.Save(&Entry{AgentName: "shared", Key: "global-fact", Content: "go has goroutines", Kind: KindFact, Scope: ScopeGlobal})

	// With limit=2 and kind=fact, both facts should be found even though
	// there are 20 pattern entries that would fill a naive limit*2 buffer.
	results, err := s.SearchWithContextOptions("agent-1", SearchOptions{
		Query: "go",
		Kind:  KindFact,
		Limit: 2,
	})
	require.NoError(t, err)
	assert.Len(t, results, 2, "should find both facts despite 20 patterns")

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Key] = true
	}
	assert.True(t, names["instance-fact"])
	assert.True(t, names["global-fact"])
}
