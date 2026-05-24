package agentmemory

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/security"
	_ "github.com/mattn/go-sqlite3"
)

func newTestEntStore(t *testing.T) *EntStore {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	return NewEntStore(client)
}

type stubPayloadProtector struct{}

func (stubPayloadProtector) EncryptPayload(plaintext []byte) ([]byte, []byte, int, error) {
	return append([]byte("enc:"), plaintext...), []byte("123456789012"), security.PayloadKeyVersionV1, nil
}

func (stubPayloadProtector) DecryptPayload(ciphertext []byte, nonce []byte, keyVersion int) ([]byte, error) {
	if keyVersion != security.PayloadKeyVersionV1 {
		return nil, errors.New("unexpected key version")
	}
	if string(nonce) != "123456789012" {
		return nil, errors.New("unexpected nonce")
	}
	return []byte(strings.TrimPrefix(string(ciphertext), "enc:")), nil
}

type failDecryptProtector struct{ stubPayloadProtector }

func (failDecryptProtector) DecryptPayload(ciphertext []byte, nonce []byte, keyVersion int) ([]byte, error) {
	return nil, errors.New("decrypt failed")
}

func TestEntStore_SaveAndGet(t *testing.T) {
	s := newTestEntStore(t)

	entry := &Entry{
		AgentName:  "researcher",
		Key:        "go-patterns",
		Content:    "Use table-driven tests",
		Kind:       KindPattern,
		Scope:      ScopeInstance,
		Confidence: 0.9,
		Tags:       []string{"go", "testing"},
	}
	require.NoError(t, s.Save(entry))

	got, err := s.Get("researcher", "go-patterns")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "go-patterns", got.Key)
	assert.Equal(t, "Use table-driven tests", got.Content)
	assert.Equal(t, KindPattern, got.Kind)
	assert.Equal(t, ScopeInstance, got.Scope)
	assert.Equal(t, 0.9, got.Confidence)
	assert.Equal(t, []string{"go", "testing"}, got.Tags)
	assert.NotEmpty(t, got.ID)
	assert.False(t, got.CreatedAt.IsZero())
	assert.False(t, got.UpdatedAt.IsZero())
}

func TestEntStore_Get_NotFound(t *testing.T) {
	s := newTestEntStore(t)

	got, err := s.Get("nobody", "nothing")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestEntStore_Save_Upsert(t *testing.T) {
	s := newTestEntStore(t)

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
}

func TestEntStore_Save_Validation(t *testing.T) {
	s := newTestEntStore(t)

	tests := []struct {
		give    *Entry
		wantErr string
	}{
		{
			give:    &Entry{Key: "no-agent", Content: "missing agent"},
			wantErr: "agent_name is required",
		},
		{
			give:    &Entry{AgentName: "a", Content: "missing key"},
			wantErr: "key is required",
		},
		{
			give: &Entry{
				AgentName: "a",
				Key:       "k",
				Content:   "c",
				Kind:      MemoryKind("bogus"),
			},
			wantErr: "invalid memory kind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantErr, func(t *testing.T) {
			err := s.Save(tt.give)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestEntStore_Search_KeywordAndSort(t *testing.T) {
	s := newTestEntStore(t)

	entries := []*Entry{
		{AgentName: "a1", Key: "go-patterns", Content: "table tests in go", Kind: KindPattern, Scope: ScopeInstance, Confidence: 0.7, Tags: []string{"go"}},
		{AgentName: "a1", Key: "go-perf", Content: "go performance tips", Kind: KindFact, Scope: ScopeInstance, Confidence: 0.9},
		{AgentName: "a1", Key: "dark-mode", Content: "user prefers dark mode", Kind: KindPreference, Scope: ScopeInstance, Confidence: 0.5},
	}
	for _, e := range entries {
		require.NoError(t, s.Save(e))
	}

	// Search for "go" — should match 2 entries, ordered by confidence DESC.
	results, err := s.Search("a1", SearchOptions{Query: "go"})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "go-perf", results[0].Key, "higher confidence first")
	assert.Equal(t, "go-patterns", results[1].Key)
}

func TestEntStore_Search_Filters(t *testing.T) {
	s := newTestEntStore(t)

	entries := []*Entry{
		{AgentName: "a1", Key: "p1", Content: "pattern one", Kind: KindPattern, Scope: ScopeInstance, Confidence: 0.9, Tags: []string{"go", "testing"}},
		{AgentName: "a1", Key: "p2", Content: "preference one", Kind: KindPreference, Scope: ScopeInstance, Confidence: 0.7},
		{AgentName: "a1", Key: "p3", Content: "low confidence fact", Kind: KindFact, Scope: ScopeInstance, Confidence: 0.2},
	}
	for _, e := range entries {
		require.NoError(t, s.Save(e))
	}

	tests := []struct {
		give      SearchOptions
		wantCount int
	}{
		{give: SearchOptions{}, wantCount: 3},
		{give: SearchOptions{Kind: KindPattern}, wantCount: 1},
		{give: SearchOptions{MinConfidence: 0.5}, wantCount: 2},
		{give: SearchOptions{Tags: []string{"go"}}, wantCount: 1},
		{give: SearchOptions{Limit: 1}, wantCount: 1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			results, err := s.Search("a1", tt.give)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
		})
	}
}

func TestEntStore_SearchWithContext_Phases(t *testing.T) {
	s := newTestEntStore(t)

	// Instance-scoped entry for agent1.
	require.NoError(t, s.Save(&Entry{
		AgentName: "agent1",
		Key:       "local-fact",
		Content:   "instance only detail",
		Kind:      KindFact,
		Scope:     ScopeInstance,
	}))

	// Global entry from agent2.
	require.NoError(t, s.Save(&Entry{
		AgentName: "agent2",
		Key:       "shared-fact",
		Content:   "global knowledge detail",
		Kind:      KindFact,
		Scope:     ScopeGlobal,
	}))

	// Instance-scoped entry from agent2 (should NOT appear for agent1).
	require.NoError(t, s.Save(&Entry{
		AgentName: "agent2",
		Key:       "private-fact",
		Content:   "agent2 private detail",
		Kind:      KindFact,
		Scope:     ScopeInstance,
	}))

	results, err := s.SearchWithContext("agent1", "detail", 10)
	require.NoError(t, err)

	// Should find: "local-fact" (instance, agent1) + "shared-fact" (global, agent2).
	// Should NOT find: "private-fact" (instance, agent2).
	assert.Len(t, results, 2)

	keys := make(map[string]bool)
	for _, r := range results {
		keys[r.Key] = true
	}
	assert.True(t, keys["local-fact"])
	assert.True(t, keys["shared-fact"])
}

func TestEntStore_SearchWithContextOptions_Filters(t *testing.T) {
	s := newTestEntStore(t)

	// Instance entries for agent-1 (different kinds).
	require.NoError(t, s.Save(&Entry{AgentName: "agent-1", Key: "local-fact", Content: "go is typed", Kind: KindFact, Scope: ScopeInstance, Confidence: 0.8}))
	require.NoError(t, s.Save(&Entry{AgentName: "agent-1", Key: "local-pattern", Content: "table driven tests", Kind: KindPattern, Scope: ScopeInstance, Confidence: 0.9}))

	// Global entry from another agent with matching kind.
	require.NoError(t, s.Save(&Entry{AgentName: "shared", Key: "global-fact", Content: "go is compiled", Kind: KindFact, Scope: ScopeGlobal, Confidence: 0.7}))

	results, err := s.SearchWithContextOptions("agent-1", SearchOptions{
		Query: "go",
		Kind:  KindFact,
		Limit: 10,
	})
	require.NoError(t, err)

	// Should include both instance fact AND global fact, but NOT the pattern.
	assert.Len(t, results, 2, "expected instance fact + global fact")
	for _, r := range results {
		assert.Equal(t, KindFact, r.Kind)
	}

	// Test MinConfidence filter.
	results, err = s.SearchWithContextOptions("agent-1", SearchOptions{
		Query:         "go",
		Kind:          KindFact,
		MinConfidence: 0.75,
		Limit:         10,
	})
	require.NoError(t, err)
	assert.Len(t, results, 1, "only the 0.8 confidence fact should match")
	assert.Equal(t, "local-fact", results[0].Key)
}

func TestEntStore_Delete(t *testing.T) {
	s := newTestEntStore(t)

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

func TestEntStore_IncrementUseCount(t *testing.T) {
	s := newTestEntStore(t)

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

func TestEntStore_Prune(t *testing.T) {
	s := newTestEntStore(t)

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

func TestEntStore_ListAgentNames(t *testing.T) {
	s := newTestEntStore(t)

	// Empty store.
	names, err := s.ListAgentNames()
	require.NoError(t, err)
	assert.Empty(t, names)

	// Add entries for multiple agents.
	require.NoError(t, s.Save(&Entry{AgentName: "beta", Key: "k1", Content: "c1", Kind: KindFact, Scope: ScopeInstance}))
	require.NoError(t, s.Save(&Entry{AgentName: "alpha", Key: "k1", Content: "c2", Kind: KindFact, Scope: ScopeInstance}))
	require.NoError(t, s.Save(&Entry{AgentName: "gamma", Key: "k1", Content: "c3", Kind: KindFact, Scope: ScopeInstance}))

	names, err = s.ListAgentNames()
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, names)
}

func TestEntStore_ListAll_Ordering(t *testing.T) {
	s := newTestEntStore(t)

	// Save entries with predictable update order.
	require.NoError(t, s.Save(&Entry{AgentName: "a1", Key: "first", Content: "c1", Kind: KindFact, Scope: ScopeInstance}))
	require.NoError(t, s.Save(&Entry{AgentName: "a1", Key: "second", Content: "c2", Kind: KindFact, Scope: ScopeInstance}))
	require.NoError(t, s.Save(&Entry{AgentName: "a1", Key: "third", Content: "c3", Kind: KindFact, Scope: ScopeInstance}))

	results, err := s.ListAll("a1")
	require.NoError(t, err)
	require.Len(t, results, 3)

	// Most recently created should be first (updated_at DESC).
	assert.Equal(t, "third", results[0].Key)
}

func TestEntStore_AgentNameIsolation(t *testing.T) {
	s := newTestEntStore(t)

	require.NoError(t, s.Save(&Entry{AgentName: "agent-a", Key: "shared-key", Content: "from agent-a", Kind: KindFact, Scope: ScopeInstance}))
	require.NoError(t, s.Save(&Entry{AgentName: "agent-b", Key: "shared-key", Content: "from agent-b", Kind: KindFact, Scope: ScopeInstance}))

	gotA, err := s.Get("agent-a", "shared-key")
	require.NoError(t, err)
	require.NotNil(t, gotA)
	assert.Equal(t, "from agent-a", gotA.Content)

	gotB, err := s.Get("agent-b", "shared-key")
	require.NoError(t, err)
	require.NotNil(t, gotB)
	assert.Equal(t, "from agent-b", gotB.Content)

	// Delete only agent-a's entry.
	require.NoError(t, s.Delete("agent-a", "shared-key"))

	gotA, err = s.Get("agent-a", "shared-key")
	require.NoError(t, err)
	assert.Nil(t, gotA)

	gotB, err = s.Get("agent-b", "shared-key")
	require.NoError(t, err)
	require.NotNil(t, gotB)
	assert.Equal(t, "from agent-b", gotB.Content)
}

func TestEntStore_PayloadProtection_UsesRedactedProjection(t *testing.T) {
	s := newTestEntStore(t)
	s.SetPayloadProtector(stubPayloadProtector{})

	entry := &Entry{
		AgentName: "researcher",
		Key:       "secret-key",
		Content:   "email alice@example.com token SECRETSECRETSECRETSECRETSECRETSECRET",
		Kind:      KindFact,
		Scope:     ScopeInstance,
	}
	require.NoError(t, s.Save(entry))

	got, err := s.Get("researcher", "secret-key")
	require.NoError(t, err)
	require.Equal(t, entry.Content, got.Content)

	row := s.client.AgentMemory.Query().OnlyX(context.Background())
	require.NotNil(t, row.ContentCiphertext)
	require.NotContains(t, row.Content, "alice@example.com")
	require.NotContains(t, row.Content, "SECRETSECRETSECRETSECRETSECRETSECRET")

	results, err := s.Search("researcher", SearchOptions{Query: "SECRETSECRETSECRETSECRETSECRETSECRET"})
	require.NoError(t, err)
	assert.Len(t, results, 0)

	results, err = s.Search("researcher", SearchOptions{Query: "secret-key"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestEntStore_ProtectedDecryptFailureDoesNotFallback(t *testing.T) {
	s := newTestEntStore(t)
	s.SetPayloadProtector(stubPayloadProtector{})
	require.NoError(t, s.Save(&Entry{
		AgentName: "researcher",
		Key:       "secret",
		Content:   "payload",
		Kind:      KindFact,
		Scope:     ScopeInstance,
	}))
	s.SetPayloadProtector(failDecryptProtector{})

	_, err := s.Get("researcher", "secret")
	require.Error(t, err)
}
