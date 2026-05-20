package search

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openFts5IndexTableLifecycleAndDbAccessorSearchDB(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "search.db")
	db, err := sql.Open("sqlite3", path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func requireFts5IndexTableLifecycleAndDbAccessorFTS5(t *testing.T, db *sql.DB) {
	t.Helper()

	if !ProbeFTS5(db) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}
}

func TestFTS5IndexTableLifecycleAndDBAccessor(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	requireFts5IndexTableLifecycleAndDbAccessorFTS5(t, db)

	idx := NewFTS5Index(db, "fts5IndexTableLifecycleAndDbAccessor_lifecycle_fts", []string{"title", "body"})
	require.Same(t, db, idx.DB())

	require.NoError(t, idx.EnsureTable())
	require.NoError(t, idx.EnsureTable())

	var createSQL string
	err := db.QueryRow(
		`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?`,
		"fts5IndexTableLifecycleAndDbAccessor_lifecycle_fts",
	).Scan(&createSQL)
	require.NoError(t, err)
	assert.Contains(t, createSQL, "CREATE VIRTUAL TABLE")
	assert.Contains(t, createSQL, "source_id UNINDEXED")
	assert.Contains(t, createSQL, "tokenize='unicode61'")

	require.NoError(t, idx.DropTable())
	require.NoError(t, idx.DropTable())

	var remaining int
	err = db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
		"fts5IndexTableLifecycleAndDbAccessor_lifecycle_fts",
	).Scan(&remaining)
	require.NoError(t, err)
	assert.Equal(t, 0, remaining)
}

func TestFTS5IndexDeterministicFallbackBranches(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	idx := NewFTS5Index(db, "fts5IndexTableLifecycleAndDbAccessor_no_extension_fts", []string{"title", "body"})

	require.Same(t, db, idx.DB())
	require.NoError(t, idx.BulkInsert(context.Background(), nil))

	results, err := idx.Search(context.Background(), "   ", 10)
	require.NoError(t, err)
	assert.Nil(t, results)

	results, err = idx.Search(context.Background(), `\ / ' ; ,`, 10)
	require.NoError(t, err)
	assert.Nil(t, results)

	bad := NewFTS5Index(db, "bad table name", []string{"title"})
	err = bad.EnsureTable()
	require.Error(t, err)
	assert.ErrorContains(t, err, "create FTS5 table bad table name")

	err = bad.DropTable()
	require.Error(t, err)
	assert.ErrorContains(t, err, "drop FTS5 table bad table name")
}

func TestFTS5IndexExecerBranchesWithoutFTS5Extension(t *testing.T) {
	ctx := context.Background()
	idx := NewFTS5Index(nil, "fts5IndexTableLifecycleAndDbAccessor_exec_fts", []string{"title", "body"})
	ex := &fts5IndexTableLifecycleAndDbAccessorExecRecorder{}

	require.NoError(t, idx.InsertWithExec(ctx, ex, "doc-1", []string{"title", "body"}))
	require.Len(t, ex.calls, 1)
	assert.Equal(t, `INSERT INTO fts5IndexTableLifecycleAndDbAccessor_exec_fts(source_id, title, body) VALUES(?, ?, ?)`, ex.calls[0].query)
	assert.Equal(t, []any{"doc-1", "title", "body"}, ex.calls[0].args)

	require.NoError(t, idx.DeleteWithExec(ctx, ex, "doc-1"))
	require.Len(t, ex.calls, 2)
	assert.Equal(t, `DELETE FROM fts5IndexTableLifecycleAndDbAccessor_exec_fts WHERE source_id = ?`, ex.calls[1].query)
	assert.Equal(t, []any{"doc-1"}, ex.calls[1].args)

	require.NoError(t, idx.UpdateWithExec(ctx, ex, "doc-2", []string{"updated", "body"}))
	require.Len(t, ex.calls, 4)
	assert.Equal(t, `DELETE FROM fts5IndexTableLifecycleAndDbAccessor_exec_fts WHERE source_id = ?`, ex.calls[2].query)
	assert.Equal(t, `INSERT INTO fts5IndexTableLifecycleAndDbAccessor_exec_fts(source_id, title, body) VALUES(?, ?, ?)`, ex.calls[3].query)

	deleteErr := errors.New("delete failed")
	ex = &fts5IndexTableLifecycleAndDbAccessorExecRecorder{err: deleteErr}
	err := idx.UpdateWithExec(ctx, ex, "doc-3", []string{"ignored", "body"})
	require.ErrorIs(t, err, deleteErr)
	assert.ErrorContains(t, err, "FTS5 delete from fts5IndexTableLifecycleAndDbAccessor_exec_fts")
	require.Len(t, ex.calls, 1)

	insertErr := errors.New("insert failed")
	ex = &fts5IndexTableLifecycleAndDbAccessorExecRecorder{errOnCall: 2, err: insertErr}
	err = idx.UpdateWithExec(ctx, ex, "doc-4", []string{"new", "body"})
	require.ErrorIs(t, err, insertErr)
	assert.ErrorContains(t, err, "FTS5 insert into fts5IndexTableLifecycleAndDbAccessor_exec_fts")
	require.Len(t, ex.calls, 2)
}

func TestFTS5IndexDirectSQLBranchesWithoutFTS5Extension(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	ctx := context.Background()
	_, err := db.Exec(`CREATE TABLE fts5IndexTableLifecycleAndDbAccessor_sql_fts(source_id TEXT PRIMARY KEY, title TEXT, body TEXT)`)
	require.NoError(t, err)
	idx := NewFTS5Index(db, "fts5IndexTableLifecycleAndDbAccessor_sql_fts", []string{"title", "body"})

	require.NoError(t, idx.Insert(ctx, "doc-1", []string{"alpha", "first body"}))
	var title string
	err = db.QueryRow(`SELECT title FROM fts5IndexTableLifecycleAndDbAccessor_sql_fts WHERE source_id = ?`, "doc-1").Scan(&title)
	require.NoError(t, err)
	assert.Equal(t, "alpha", title)

	require.NoError(t, idx.Update(ctx, "doc-1", []string{"beta", "updated body"}))
	err = db.QueryRow(`SELECT title FROM fts5IndexTableLifecycleAndDbAccessor_sql_fts WHERE source_id = ?`, "doc-1").Scan(&title)
	require.NoError(t, err)
	assert.Equal(t, "beta", title)

	require.NoError(t, idx.Delete(ctx, "doc-1"))
	var count int
	err = db.QueryRow(`SELECT count(*) FROM fts5IndexTableLifecycleAndDbAccessor_sql_fts`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	require.NoError(t, idx.BulkInsert(ctx, []Record{
		{RowID: "doc-2", Values: []string{"bulk alpha", "shared body"}},
		{RowID: "doc-3", Values: []string{"bulk beta", "shared body"}},
	}))
	err = db.QueryRow(`SELECT count(*) FROM fts5IndexTableLifecycleAndDbAccessor_sql_fts`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	err = idx.BulkInsert(ctx, []Record{
		{RowID: "doc-4", Values: []string{"valid", "before rollback"}},
		{RowID: "doc-5", Values: []string{"missing body"}},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "FTS5 bulk insert row doc-5")
	err = db.QueryRow(`SELECT count(*) FROM fts5IndexTableLifecycleAndDbAccessor_sql_fts`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	_, err = idx.Search(ctx, "alpha", 1)
	require.Error(t, err)
	assert.ErrorContains(t, err, "FTS5 search fts5IndexTableLifecycleAndDbAccessor_sql_fts")
}

func TestFTS5IndexInsertUpdateDeleteWithTransactionExec(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	requireFts5IndexTableLifecycleAndDbAccessorFTS5(t, db)

	ctx := context.Background()
	idx := NewFTS5Index(db, "fts5IndexTableLifecycleAndDbAccessor_tx_fts", []string{"title", "body"})
	require.NoError(t, idx.EnsureTable())

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, idx.InsertWithExec(ctx, tx, "doc-1", []string{
		"original note",
		"alpha content before update",
	}))
	require.NoError(t, idx.UpdateWithExec(ctx, tx, "doc-1", []string{
		"updated note",
		"omega content after update",
	}))
	require.NoError(t, tx.Commit())

	results, err := idx.Search(ctx, "omega", 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "doc-1", results[0].RowID)

	results, err = idx.Search(ctx, "alpha", 10)
	require.NoError(t, err)
	assert.Empty(t, results)

	tx, err = db.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, idx.DeleteWithExec(ctx, tx, "doc-1"))
	require.NoError(t, tx.Commit())

	results, err = idx.Search(ctx, "omega", 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestFTS5IndexDirectInsertUpdateDeleteAndSearchLimits(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	requireFts5IndexTableLifecycleAndDbAccessorFTS5(t, db)

	ctx := context.Background()
	idx := NewFTS5Index(db, "fts5IndexTableLifecycleAndDbAccessor_direct_fts", []string{"title", "body"})
	require.NoError(t, idx.EnsureTable())

	require.NoError(t, idx.Insert(ctx, "doc-1", []string{"deploy guide", "deploy safely"}))
	require.NoError(t, idx.Insert(ctx, "doc-2", []string{"deploy checklist", "deploy repeatedly"}))
	require.NoError(t, idx.Insert(ctx, "doc-3", []string{"database guide", "configure storage"}))

	results, err := idx.Search(ctx, "deploy", 0)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.ElementsMatch(t, []string{"doc-1", "doc-2"}, resultRowIDs(results))

	results, err = idx.Search(ctx, "???", 10)
	require.NoError(t, err)
	assert.Nil(t, results)

	require.NoError(t, idx.Update(ctx, "doc-2", []string{"rollback checklist", "restore service"}))
	results, err = idx.Search(ctx, "deploy", -5)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "doc-1", results[0].RowID)

	require.NoError(t, idx.Delete(ctx, "doc-1"))
	results, err = idx.Search(ctx, "deploy", 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestFTS5IndexBulkInsertSuccessAndRollbackOnRowError(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	requireFts5IndexTableLifecycleAndDbAccessorFTS5(t, db)

	ctx := context.Background()
	idx := NewFTS5Index(db, "fts5IndexTableLifecycleAndDbAccessor_bulk_fts", []string{"title", "body"})
	require.NoError(t, idx.EnsureTable())

	err := idx.BulkInsert(ctx, []Record{
		{RowID: "bad-1", Values: []string{"first valid", "rollback marker"}},
		{RowID: "bad-2", Values: []string{"missing body"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FTS5 bulk insert row bad-2")

	results, searchErr := idx.Search(ctx, "rollback", 10)
	require.NoError(t, searchErr)
	assert.Empty(t, results)

	require.NoError(t, idx.BulkInsert(ctx, []Record{
		{RowID: "doc-1", Values: []string{"alpha", "shared topic"}},
		{RowID: "doc-2", Values: []string{"beta", "shared topic"}},
		{RowID: "doc-3", Values: []string{"gamma", "isolated topic"}},
	}))

	results, err = idx.Search(ctx, "shared", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, []string{"doc-1", "doc-2"}, results[0].RowID)
	require.NoError(t, idx.BulkInsert(ctx, nil))
}

func TestFTS5IndexErrorBranchesWrapOperationContext(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	requireFts5IndexTableLifecycleAndDbAccessorFTS5(t, db)

	ctx := context.Background()
	idx := NewFTS5Index(db, "fts5IndexTableLifecycleAndDbAccessor_errors_fts", []string{"title", "body"})

	_, err := idx.Search(ctx, "anything", 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FTS5 search fts5IndexTableLifecycleAndDbAccessor_errors_fts")

	err = idx.DropTable()
	require.NoError(t, err)

	bad := NewFTS5Index(db, "bad table name", []string{"title"})
	err = bad.EnsureTable()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create FTS5 table bad table name")

	err = bad.DropTable()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "drop FTS5 table bad table name")

	require.NoError(t, idx.EnsureTable())
	err = idx.Insert(ctx, "bad-values", []string{"too few values"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FTS5 insert into fts5IndexTableLifecycleAndDbAccessor_errors_fts")
}

func TestFTS5IndexSearchEscapesProblematicQueries(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	requireFts5IndexTableLifecycleAndDbAccessorFTS5(t, db)

	ctx := context.Background()
	idx := NewFTS5Index(db, "fts5IndexTableLifecycleAndDbAccessor_escape_fts", []string{"title", "body"})
	require.NoError(t, idx.EnsureTable())
	require.NoError(t, idx.BulkInsert(ctx, []Record{
		{RowID: "doc-1", Values: []string{"quoted phrase", "alpha beta gamma"}},
		{RowID: "doc-2", Values: []string{"prefix token", "deployment deployable deployed"}},
		{RowID: "doc-3", Values: []string{"sanitized token", "email address supportexamplecom"}},
	}))

	tests := []struct {
		name string
		give string
		want []string
	}{
		{name: "phrase", give: `"alpha beta"`, want: []string{"doc-1"}},
		{name: "prefix", give: "deploy*", want: []string{"doc-2"}},
		{name: "unclosed quote becomes token", give: `"gamma`, want: []string{"doc-1"}},
		{name: "punctuation is stripped", give: "support@example.com", want: []string{"doc-3"}},
		{name: "boolean-looking punctuation is inert", give: "alpha + beta - gamma", want: []string{"doc-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := idx.Search(ctx, tt.give, 10)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, resultRowIDs(results))
		})
	}
}

func TestSanitizeFTS5QueryBranches(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{name: "tabs become separators", give: "alpha\tbeta", want: "alpha OR beta"},
		{name: "quoted phrase followed by token", give: `"alpha beta" gamma`, want: `"alpha beta" OR gamma`},
		{name: "token stops before quote", give: `alpha"beta gamma"`, want: `alpha OR "beta gamma"`},
		{name: "unclosed quote strips punctuation", give: `"support@example.com`, want: "supportexamplecom"},
		{name: "prefix strips unsafe punctuation", give: "deploy-*", want: "deploy*"},
		{name: "all escaped token disappears", give: `\ / ' ; ,`, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeFTS5Query(tt.give))
		})
	}
}

func resultRowIDs(results []SearchResult) []string {
	ids := make([]string, 0, len(results))
	for _, result := range results {
		ids = append(ids, result.RowID)
	}
	return ids
}

type fts5IndexTableLifecycleAndDbAccessorExecCall struct {
	query string
	args  []any
}

type fts5IndexTableLifecycleAndDbAccessorExecRecorder struct {
	calls     []fts5IndexTableLifecycleAndDbAccessorExecCall
	errOnCall int
	err       error
}

func (e *fts5IndexTableLifecycleAndDbAccessorExecRecorder) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	e.calls = append(e.calls, fts5IndexTableLifecycleAndDbAccessorExecCall{
		query: query,
		args:  append([]any(nil), args...),
	})
	if e.err != nil && (e.errOnCall == 0 || e.errOnCall == len(e.calls)) {
		return nil, e.err
	}
	return fts5IndexTableLifecycleAndDbAccessorSQLResult(1), nil
}

type fts5IndexTableLifecycleAndDbAccessorSQLResult int64

func (r fts5IndexTableLifecycleAndDbAccessorSQLResult) LastInsertId() (int64, error) {
	return int64(r), nil
}
func (r fts5IndexTableLifecycleAndDbAccessorSQLResult) RowsAffected() (int64, error) {
	return int64(r), nil
}

func TestFTS5IndexUsesTempFileSQLite(t *testing.T) {
	db := openFts5IndexTableLifecycleAndDbAccessorSearchDB(t)
	requireFts5IndexTableLifecycleAndDbAccessorFTS5(t, db)

	var rows []string
	values, err := db.Query(`PRAGMA database_list`)
	require.NoError(t, err)
	defer values.Close()

	for values.Next() {
		var seq int
		var name, file string
		require.NoError(t, values.Scan(&seq, &name, &file))
		if name == "main" {
			rows = append(rows, file)
		}
	}
	require.NoError(t, values.Err())
	require.Len(t, rows, 1)
	assert.True(t, strings.HasSuffix(rows[0], "search.db"), rows[0])
}
