package search

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func skipWithoutFTS5(t *testing.T, db *sql.DB) {
	t.Helper()
	if !ProbeFTS5(db) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}
}

func TestProbeFTS5(t *testing.T) {
	db := openTestDB(t)
	// ProbeFTS5 should not panic regardless of availability.
	got := ProbeFTS5(db)
	t.Logf("FTS5 available: %v", got)
}

func TestProbeFTS5_Idempotent(t *testing.T) {
	db := openTestDB(t)
	first := ProbeFTS5(db)
	second := ProbeFTS5(db)
	assert.Equal(t, first, second)
}

func TestProbeFTS5_NoResidualTables(t *testing.T) {
	db := openTestDB(t)

	ProbeFTS5(db)

	var count int
	err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name = '_fts5_probe'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestFTS5Index_TableLifecycle(t *testing.T) {
	db := openTestDB(t)
	skipWithoutFTS5(t, db)

	idx := NewFTS5Index(db, "test_fts", []string{"title", "body"})

	t.Run("create table", func(t *testing.T) {
		require.NoError(t, idx.EnsureTable())
	})

	t.Run("create again is idempotent", func(t *testing.T) {
		require.NoError(t, idx.EnsureTable())
	})

	t.Run("drop table", func(t *testing.T) {
		require.NoError(t, idx.DropTable())
	})

	t.Run("drop non-existent is ok", func(t *testing.T) {
		require.NoError(t, idx.DropTable())
	})
}

func TestFTS5Index_CRUD(t *testing.T) {
	db := openTestDB(t)
	skipWithoutFTS5(t, db)
	idx := NewFTS5Index(db, "crud_fts", []string{"key", "content"})
	require.NoError(t, idx.EnsureTable())
	ctx := context.Background()

	t.Run("insert and search", func(t *testing.T) {
		require.NoError(t, idx.Insert(ctx, "k1", []string{"deploy", "deploy server to production"}))
		require.NoError(t, idx.Insert(ctx, "k2", []string{"config", "configuration file for database"}))

		results, err := idx.Search(ctx, "deploy", 10)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "k1", results[0].RowID)
	})

	t.Run("update changes content", func(t *testing.T) {
		require.NoError(t, idx.Update(ctx, "k1", []string{"deploy", "updated deployment guide"}))

		results, err := idx.Search(ctx, "guide", 10)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "k1", results[0].RowID)

		// Old content should not match.
		results, err = idx.Search(ctx, "production", 10)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("delete removes entry", func(t *testing.T) {
		require.NoError(t, idx.Delete(ctx, "k2"))

		results, err := idx.Search(ctx, "configuration", 10)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestFTS5Index_BulkInsert(t *testing.T) {
	db := openTestDB(t)
	skipWithoutFTS5(t, db)
	idx := NewFTS5Index(db, "bulk_fts", []string{"title", "body"})
	require.NoError(t, idx.EnsureTable())
	ctx := context.Background()

	records := []Record{
		{RowID: "r1", Values: []string{"alpha", "first entry about alpha"}},
		{RowID: "r2", Values: []string{"beta", "second entry about beta"}},
		{RowID: "r3", Values: []string{"gamma", "third entry about gamma"}},
	}

	t.Run("bulk insert", func(t *testing.T) {
		require.NoError(t, idx.BulkInsert(ctx, records))

		results, err := idx.Search(ctx, "beta", 10)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "r2", results[0].RowID)
	})

	t.Run("bulk insert empty is noop", func(t *testing.T) {
		require.NoError(t, idx.BulkInsert(ctx, nil))
	})
}

func TestFTS5Index_Search(t *testing.T) {
	db := openTestDB(t)
	skipWithoutFTS5(t, db)
	idx := NewFTS5Index(db, "search_fts", []string{"key", "content"})
	require.NoError(t, idx.EnsureTable())
	ctx := context.Background()

	// Seed data.
	records := []Record{
		{RowID: "k1", Values: []string{"deploy-guide", "how to deploy a server to production"}},
		{RowID: "k2", Values: []string{"deploy-rollback", "rolling back a failed deployment"}},
		{RowID: "k3", Values: []string{"database-config", "configure database connection pool"}},
		{RowID: "k4", Values: []string{"server-setup", "initial server setup and deployment steps"}},
	}
	require.NoError(t, idx.BulkInsert(ctx, records))

	tests := []struct {
		give      string
		wantCount int
		wantFirst string
	}{
		{give: "deploy", wantCount: 2},                  // k1, k2 contain "deploy" (k4 has "deployment" — different token)
		{give: "database", wantCount: 1, wantFirst: "k3"},
		{give: `"deploy a server"`, wantCount: 1, wantFirst: "k1"}, // phrase match
		{give: "dep*", wantCount: 3},                                // prefix match
		{give: "", wantCount: 0},                                    // empty query
		{give: "nonexistent", wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			results, err := idx.Search(ctx, tt.give, 10)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
			if tt.wantFirst != "" && len(results) > 0 {
				assert.Equal(t, tt.wantFirst, results[0].RowID)
			}
		})
	}
}

func TestFTS5Index_SearchLimit(t *testing.T) {
	db := openTestDB(t)
	skipWithoutFTS5(t, db)
	idx := NewFTS5Index(db, "limit_fts", []string{"body"})
	require.NoError(t, idx.EnsureTable())
	ctx := context.Background()

	for i := 0; i < 20; i++ {
		require.NoError(t, idx.Insert(ctx, fmt.Sprintf("r%d", i), []string{"common keyword here"}))
	}

	results, err := idx.Search(ctx, "keyword", 5)
	require.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestFTS5Index_ConcurrentAccess(t *testing.T) {
	db := openTestDB(t)
	skipWithoutFTS5(t, db)
	idx := NewFTS5Index(db, "concurrent_fts", []string{"body"})
	require.NoError(t, idx.EnsureTable())
	ctx := context.Background()

	// Seed some data.
	for i := 0; i < 10; i++ {
		require.NoError(t, idx.Insert(ctx, fmt.Sprintf("c%d", i), []string{fmt.Sprintf("entry number %d about topics", i)}))
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = idx.Search(ctx, "topics", 5)
		}()
	}
	wg.Wait()
}

func TestSanitizeFTS5Query(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "", want: ""},
		{give: "  ", want: ""},
		{give: "hello", want: "hello"},
		{give: "hello world", want: "hello OR world"},
		{give: `"exact phrase"`, want: `"exact phrase"`},
		{give: "dep*", want: "dep*"},
		{give: `hello "exact phrase" world`, want: `hello OR "exact phrase" OR world`},
		{give: "special:chars", want: "specialchars"},
		// Punctuation-only queries should produce empty result.
		{give: "?", want: ""},
		{give: ".", want: ""},
		{give: "!", want: ""},
		{give: "???", want: ""},
		// Mixed punctuation and text.
		{give: "hello? world.", want: "hello OR world"},
		// Prefix token with only punctuation.
		{give: "?*", want: ""},
		// Unclosed quote with only punctuation.
		{give: `"?`, want: ""},
		{give: `"!@#`, want: ""},
		// Korean text should pass through.
		{give: "안녕하세요", want: "안녕하세요"},
		{give: "한국어 테스트", want: "한국어 OR 테스트"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := sanitizeFTS5Query(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
