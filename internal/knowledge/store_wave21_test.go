package knowledge

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/enttest"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/sqlitedriver"
)

func TestWave21SaveKnowledgeAtomicRollsBackWhenFTS5InsertFails(t *testing.T) {
	store, rawDB := newWave21FTS5TestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()
	require.NoError(t, store.fts5Index.DropTable())

	err := store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave21-knowledge-rollback",
		Category: entknowledge.CategoryFact,
		Content:  "atomic insert should roll back when fts5 is unavailable",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "FTS5 insert")

	_, err = store.GetKnowledge(ctx, "wave21-knowledge-rollback")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))

	var tableName string
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'knowledge_fts'`,
	).Scan(&tableName)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestWave21SaveLearningAtomicRollsBackWhenFTS5InsertFails(t *testing.T) {
	store, rawDB := newWave21FTS5TestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()
	require.NoError(t, store.learningFTS5Idx.DropTable())

	err := store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave21 learning rollback",
		ErrorPattern: "atomic learning insert should fail",
		Diagnosis:    "fts5 table is unavailable",
		Fix:          "rollback ent row",
		Category:     entlearning.CategoryToolError,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "sync learning FTS5")
	require.Contains(t, err.Error(), "FTS5 insert")

	rows, err := store.client.Learning.Query().All(ctx)
	require.NoError(t, err)
	require.Empty(t, rows)

	var tableName string
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'learning_fts'`,
	).Scan(&tableName)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestWave21SearchKnowledgeScoredUsesFTS5RankAndFiltersCategory(t *testing.T) {
	store, _ := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave21-scored-fact",
		Category: entknowledge.CategoryFact,
		Content:  "wave21 ranked needle content",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave21-scored-rule",
		Category: entknowledge.CategoryRule,
		Content:  "wave21 ranked needle content",
	}))

	got, err := store.SearchKnowledgeScored(
		ctx,
		"needle",
		string(entknowledge.CategoryFact),
		10,
	)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "wave21-scored-fact", got[0].Entry.Key)
	require.Equal(t, "fts5", got[0].SearchSource)
	require.Greater(t, got[0].Score, 0.0)
}

func TestWave21SyncLearningFTS5WithExecBranches(t *testing.T) {
	store, rawDB := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.syncLearningFTS5WithExec(
		ctx,
		rawDB,
		"wave21-learning-sync",
		"initial trigger",
		"initial error",
		"initial fix",
		false,
	))

	var count int
	err := rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM learning_fts WHERE source_id = ? AND learning_fts MATCH 'initial'`,
		"wave21-learning-sync",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	require.NoError(t, store.syncLearningFTS5WithExec(
		ctx,
		rawDB,
		"wave21-learning-sync",
		"updated trigger",
		"updated error",
		"updated fix",
		true,
	))

	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM learning_fts WHERE source_id = ? AND learning_fts MATCH 'updated'`,
		"wave21-learning-sync",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM learning_fts WHERE source_id = ? AND learning_fts MATCH 'initial'`,
		"wave21-learning-sync",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	require.NoError(t, store.learningFTS5Idx.DropTable())
	err = store.syncLearningFTS5WithExec(
		ctx,
		rawDB,
		"wave21-learning-sync",
		"broken trigger",
		"broken error",
		"broken fix",
		true,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sync learning FTS5 wave21-learning-sync")

	store.SetLearningFTS5Index(nil)
	require.NoError(t, store.syncLearningFTS5WithExec(
		ctx,
		rawDB,
		"wave21-learning-sync",
		"ignored trigger",
		"ignored error",
		"ignored fix",
		false,
	))
}

func TestWave21PrepareLearningWriteProtectsJSONAndRedactsProjection(t *testing.T) {
	store := newTestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})

	projection, ciphertext, nonce, keyVersion, err := store.prepareLearningWrite(LearningEntry{
		ErrorPattern: "account alice@example.com failed with token SECRETSECRETSECRETSECRETSECRETSECRET",
		Diagnosis:    "customer 1234567890 exposed in diagnostic",
		Fix:          "rotate token SECRETSECRETSECRETSECRETSECRETSECRET",
	})
	require.NoError(t, err)
	require.NotEmpty(t, ciphertext)
	require.NotEmpty(t, nonce)
	require.Equal(t, 1, keyVersion)
	require.NotContains(t, projection.ErrorPattern, "alice@example.com")
	require.NotContains(t, projection.ErrorPattern, "SECRETSECRETSECRETSECRETSECRETSECRET")
	require.NotContains(t, projection.Diagnosis, "1234567890")
	require.NotContains(t, projection.Fix, "SECRETSECRETSECRETSECRETSECRETSECRET")
}

func TestWave21SyncLearningFTS5NoopsWhenIndexIsNil(t *testing.T) {
	store := newTestStore(t)

	require.NotPanics(t, func() {
		store.syncLearningFTS5(context.Background(), "missing", "trigger", "error", "fix")
		store.deleteLearningFTS5(context.Background(), "missing")
	})
}

func TestWave21FTS5SearchReturnsNoRowsForInvalidQueryWithoutFallback(t *testing.T) {
	store, _ := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave21-invalid-query",
		Category: entknowledge.CategoryFact,
		Content:  "plain content for invalid fts query",
	}))

	got, err := store.SearchKnowledge(ctx, "\"", "", 10)
	require.NoError(t, err)
	require.Empty(t, got)

	scored, err := store.SearchKnowledgeScored(ctx, "\"", "", 10)
	require.NoError(t, err)
	require.Empty(t, scored)
}

func TestWave21SyncLearningFTS5WithExecNilIndexDoesNotRequireFTS5(t *testing.T) {
	store := newTestStore(t)

	err := store.syncLearningFTS5WithExec(
		context.Background(),
		nilExec{},
		"wave21-nil-index",
		"trigger",
		"error",
		"fix",
		true,
	)
	require.NoError(t, err)
}

type nilExec struct{}

func (nilExec) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, errors.New("should not be called")
}

func newWave21FTS5TestStore(t *testing.T) (*Store, *sql.DB) {
	t.Helper()

	dsn := sqlitedriver.MemoryDSN("wave21-knowledge-fts")
	rawDB, err := sql.Open(sqlitedriver.DriverName(), dsn)
	require.NoError(t, err)
	require.NoError(t, sqlitedriver.ConfigureConnection(rawDB, false))
	_, err = rawDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	require.True(t, search.ProbeFTS5(rawDB), "modernc SQLite test driver must expose FTS5 for Wave 21 coverage")

	client := enttest.NewClient(t, enttest.WithOptions(ent.Driver(entsql.OpenDB(dialect.SQLite, rawDB))))
	t.Cleanup(func() { client.Close() })

	store := NewStore(client, zap.NewNop().Sugar())

	knowledgeIdx := search.NewFTS5Index(rawDB, "knowledge_fts", []string{"key", "content"})
	require.NoError(t, knowledgeIdx.EnsureTable())
	store.SetFTS5Index(knowledgeIdx)

	learningIdx := search.NewFTS5Index(rawDB, "learning_fts", []string{"trigger", "error_pattern", "fix"})
	require.NoError(t, learningIdx.EnsureTable())
	store.SetLearningFTS5Index(learningIdx)

	return store, rawDB
}
