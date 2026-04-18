package knowledge

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	_ "github.com/mattn/go-sqlite3"

	"github.com/langoai/lango/internal/search"
)

// newFTS5TestStore creates a store with both Ent and FTS5 indexes wired up.
// Skips if FTS5 is unavailable in the current SQLite runtime.
func newFTS5TestStore(t *testing.T) (*Store, *sql.DB) {
	t.Helper()

	// Ent client for ORM operations.
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	// Separate raw DB for FTS5 (in-memory, shared cache so Ent and FTS5 coexist).
	rawDB, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { rawDB.Close() })

	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	logger := zap.NewNop().Sugar()
	store := NewStore(client, logger)

	// Create and inject FTS5 indexes.
	knowledgeIdx := search.NewFTS5Index(rawDB, "knowledge_fts", []string{"key", "content"})
	require.NoError(t, knowledgeIdx.EnsureTable())
	store.SetFTS5Index(knowledgeIdx)

	learningIdx := search.NewFTS5Index(rawDB, "learning_fts", []string{"trigger", "error_pattern", "fix"})
	require.NoError(t, learningIdx.EnsureTable())
	store.SetLearningFTS5Index(learningIdx)

	return store, rawDB
}

func TestSearchKnowledge_FTS5Path(t *testing.T) {
	store, _ := newFTS5TestStore(t)
	ctx := context.Background()

	// Seed knowledge entries.
	entries := []KnowledgeEntry{
		{Key: "deploy-guide", Category: entknowledge.CategoryFact, Content: "how to deploy a server to production"},
		{Key: "db-config", Category: entknowledge.CategoryDefinition, Content: "configure database connection pool settings"},
		{Key: "deploy-rollback", Category: entknowledge.CategoryPattern, Content: "rolling back a failed deployment process"},
	}
	for _, e := range entries {
		require.NoError(t, store.SaveKnowledge(ctx, "s1", e))
	}

	tests := []struct {
		give         string
		giveCategory string
		wantCount    int
		wantFirst    string
	}{
		{give: "deploy", wantCount: 2},
		{give: "database", wantCount: 1, wantFirst: "db-config"},
		{give: "deploy", giveCategory: string(entknowledge.CategoryFact), wantCount: 1, wantFirst: "deploy-guide"},
		{give: "nonexistent", wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.give+"_cat="+tt.giveCategory, func(t *testing.T) {
			results, err := store.SearchKnowledge(ctx, tt.give, tt.giveCategory, 10)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
			if tt.wantFirst != "" && len(results) > 0 {
				assert.Equal(t, tt.wantFirst, results[0].Key)
			}
		})
	}
}

func TestSearchKnowledge_LIKEFallback(t *testing.T) {
	// Store without FTS5 index — should use LIKE path.
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	store := NewStore(client, zap.NewNop().Sugar())

	ctx := context.Background()
	require.NoError(t, store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
		Key: "go-style", Category: entknowledge.CategoryRule, Content: "Use gofmt for formatting",
	}))

	results, err := store.SearchKnowledge(ctx, "gofmt", "", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "go-style", results[0].Key)
}

func TestSearchLearnings_FTS5Path(t *testing.T) {
	store, _ := newFTS5TestStore(t)
	ctx := context.Background()

	// Seed learning entries.
	require.NoError(t, store.SaveLearning(ctx, "s1", LearningEntry{
		Trigger:      "connection refused",
		ErrorPattern: "dial tcp 127.0.0.1:5432: connection refused",
		Fix:          "ensure PostgreSQL service is running",
		Category:     entlearning.CategoryToolError,
	}))
	require.NoError(t, store.SaveLearning(ctx, "s1", LearningEntry{
		Trigger:      "permission denied",
		ErrorPattern: "open /etc/config: permission denied",
		Fix:          "run with elevated privileges",
		Category:     entlearning.CategoryPermission,
	}))

	results, err := store.SearchLearnings(ctx, "connection refused", "", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "connection refused", results[0].Trigger)
}

func TestWriteTimeSync_Knowledge(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	ctx := context.Background()

	// Insert v1 via SaveKnowledge → should appear in FTS5.
	require.NoError(t, store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
		Key: "sync-test", Category: entknowledge.CategoryFact, Content: "original content about deployment",
	}))

	var count int
	err := rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts WHERE knowledge_fts MATCH 'deployment'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Append v2 via SaveKnowledge → FTS5 should reflect new content only.
	require.NoError(t, store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
		Key: "sync-test", Category: entknowledge.CategoryFact, Content: "updated content about configuration",
	}))

	err = rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts WHERE knowledge_fts MATCH 'configuration'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Old content should not match (FTS5 updated to latest version only).
	err = rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts WHERE knowledge_fts MATCH 'deployment'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Verify only 1 FTS5 row despite 2 DB rows.
	err = rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "FTS5 should have exactly 1 row per key (latest only)")

	// Delete via DeleteKnowledge → should disappear from FTS5.
	require.NoError(t, store.DeleteKnowledge(ctx, "sync-test"))

	err = rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts WHERE knowledge_fts MATCH 'configuration'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestFTS5_OnlyLatestVersion(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	ctx := context.Background()

	// Save 3 versions with distinct content.
	for i, content := range []string{"alpha unique v1", "beta unique v2", "gamma unique v3"} {
		require.NoError(t, store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
			Key: "fts5-latest", Category: entknowledge.CategoryFact, Content: content,
		}), "save v%d", i+1)
	}

	// FTS5 should contain only v3 content.
	var count int
	err := rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts WHERE knowledge_fts MATCH 'gamma'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "FTS5 should find v3 content 'gamma'")

	// v1 and v2 content should NOT be in FTS5.
	err = rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts WHERE knowledge_fts MATCH 'alpha'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "FTS5 should NOT find v1 content 'alpha'")

	err = rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts WHERE knowledge_fts MATCH 'beta'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "FTS5 should NOT find v2 content 'beta'")

	// Total FTS5 rows for this key should be 1.
	err = rawDB.QueryRow(`SELECT count(*) FROM knowledge_fts WHERE source_id = 'fts5-latest'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "FTS5 should have exactly 1 row for this key")
}

func TestWriteTimeSync_Learning(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "s1", LearningEntry{
		Trigger:      "timeout error",
		ErrorPattern: "context deadline exceeded alice@example.com",
		Fix:          "increase timeout to 30s token SECRETSECRETSECRETSECRETSECRETSECRET",
		Category:     entlearning.CategoryTimeout,
	}))

	var count int
	err := rawDB.QueryRow(`SELECT count(*) FROM learning_fts WHERE learning_fts MATCH 'timeout'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var trigger, errorPattern, fix string
	err = rawDB.QueryRow(`SELECT trigger, error_pattern, fix FROM learning_fts LIMIT 1`).Scan(&trigger, &errorPattern, &fix)
	require.NoError(t, err)
	assert.NotContains(t, errorPattern, "alice@example.com")
	assert.NotContains(t, fix, "SECRETSECRETSECRETSECRETSECRETSECRET")
}

func TestSearchKnowledge_FTS5ErrorFallback(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	rawDB, err := sql.Open("sqlite3", "file:ent?mode=memory&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { rawDB.Close() })

	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available")
	}

	store := NewStore(client, zap.NewNop().Sugar())
	ctx := context.Background()

	// Seed data via LIKE path first.
	require.NoError(t, store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
		Key: "fallback-test", Category: entknowledge.CategoryFact, Content: "content for fallback testing",
	}))

	// Inject a broken FTS5 index (table dropped after injection).
	brokenIdx := search.NewFTS5Index(rawDB, "broken_fts", []string{"key", "content"})
	require.NoError(t, brokenIdx.EnsureTable())
	store.SetFTS5Index(brokenIdx)
	require.NoError(t, brokenIdx.DropTable()) // break it

	// Search should degrade to LIKE path and still return results.
	results, err := store.SearchKnowledge(ctx, "fallback", "", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "fallback-test", results[0].Key)
}
