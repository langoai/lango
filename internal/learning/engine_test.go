package learning

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/knowledge"
	_ "github.com/mattn/go-sqlite3"
)

func newTestEngine(t *testing.T) (*Engine, *knowledge.Store) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	store := knowledge.NewStore(client, logger)
	engine := NewEngine(store, logger)
	return engine, store
}

func TestEngine_OnToolResult_Success(t *testing.T) {
	t.Parallel()

	engine, _ := newTestEngine(t)
	ctx := context.Background()

	// Calling OnToolResult with nil error should not panic and should save an audit log.
	engine.OnToolResult(ctx, "sess-1", "file_read", map[string]interface{}{"path": "/tmp"}, "ok", nil)
}

func TestEngine_OnToolResult_Error_NewPattern(t *testing.T) {
	t.Parallel()

	engine, store := newTestEngine(t)
	ctx := context.Background()

	testErr := errors.New("connection refused")
	engine.OnToolResult(ctx, "sess-1", "http_call", nil, nil, testErr)

	// Verify a new learning was created by searching for the error pattern.
	learnings, err := store.SearchLearnings(ctx, "connection refused", "", 10)
	require.NoError(t, err)
	require.NotEmpty(t, learnings, "expected at least one learning after OnToolResult with error")

	found := false
	for _, l := range learnings {
		if l.Trigger == "tool:http_call" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected learning with trigger %q", "tool:http_call")
}

func TestEngine_OnToolResult_Error_KnownFix(t *testing.T) {
	t.Parallel()

	engine, store := newTestEngine(t)
	ctx := context.Background()

	// Create a learning with a fix and set confidence > 0.5 using the ent client directly.
	err := store.SaveLearning(ctx, "sess-1", knowledge.LearningEntry{
		Trigger:      "tool:http_call",
		ErrorPattern: "connection refused",
		Diagnosis:    "server is down",
		Fix:          "restart the server",
		Category:     entlearning.CategoryToolError,
	})
	require.NoError(t, err)

	// Boost confidence above 0.5 by searching and updating directly.
	entities, err := store.SearchLearningEntities(ctx, "connection refused", 5)
	require.NoError(t, err)
	require.NotEmpty(t, entities, "expected at least one entity")

	// Set confidence to 0.8 directly via ent update.
	_, err = entities[0].Update().SetConfidence(0.8).SetSuccessCount(10).Save(ctx)
	require.NoError(t, err)

	// Count learnings before calling OnToolResult with a matching error.
	before, err := store.SearchLearnings(ctx, "connection refused", "", 50)
	require.NoError(t, err)
	beforeCount := len(before)

	// Call OnToolResult with matching error - should NOT create a new learning
	// because a high-confidence fix already exists.
	testErr := errors.New("connection refused")
	engine.OnToolResult(ctx, "sess-2", "http_call", nil, nil, testErr)

	after, err := store.SearchLearnings(ctx, "connection refused", "", 50)
	require.NoError(t, err)
	assert.Equal(t, beforeCount, len(after), "expected no new learning")
}

func TestEngine_GetFixForError(t *testing.T) {
	t.Parallel()

	engine, store := newTestEngine(t)
	ctx := context.Background()

	t.Run("returns fix for high-confidence learning", func(t *testing.T) {
		errMsg := "undefined variable in scope"
		err := store.SaveLearning(ctx, "sess-1", knowledge.LearningEntry{
			Trigger:      "tool:compile",
			ErrorPattern: errMsg,
			Diagnosis:    "missing declaration",
			Fix:          "declare the variable before use",
			Category:     entlearning.CategoryToolError,
		})
		require.NoError(t, err)

		// Set confidence above 0.5.
		entities, err := store.SearchLearningEntities(ctx, errMsg, 5)
		require.NoError(t, err)
		require.NotEmpty(t, entities, "expected at least one entity")
		_, err = entities[0].Update().SetConfidence(0.8).Save(ctx)
		require.NoError(t, err)

		fix, ok := engine.GetFixForError(ctx, "compile", errors.New(errMsg))
		require.True(t, ok)
		assert.Equal(t, "declare the variable before use", fix)
	})

	t.Run("returns false for non-matching error", func(t *testing.T) {
		fix, ok := engine.GetFixForError(ctx, "compile", errors.New("completely unrelated xyz error"))
		assert.False(t, ok, "GetFixForError returned true for non-matching error")
		assert.Empty(t, fix)
	})

	t.Run("returns false for low-confidence learning", func(t *testing.T) {
		err := store.SaveLearning(ctx, "sess-2", knowledge.LearningEntry{
			Trigger:      "tool:deploy",
			ErrorPattern: "low conf pattern xyz",
			Diagnosis:    "some diagnosis",
			Fix:          "some fix",
			Category:     entlearning.CategoryToolError,
		})
		require.NoError(t, err)

		// Set confidence below 0.5.
		entities, err := store.SearchLearningEntities(ctx, "low conf pattern xyz", 5)
		require.NoError(t, err)
		require.NotEmpty(t, entities, "expected at least one entity")
		_, err = entities[0].Update().SetConfidence(0.3).Save(ctx)
		require.NoError(t, err)

		fix, ok := engine.GetFixForError(ctx, "deploy", errors.New("low conf pattern xyz"))
		assert.False(t, ok, "GetFixForError returned true for low-confidence learning")
		assert.Empty(t, fix)
	})
}

func TestEngine_RecordUserCorrection(t *testing.T) {
	t.Parallel()

	engine, store := newTestEngine(t)
	ctx := context.Background()

	err := engine.RecordUserCorrection(ctx, "sess-1", "wrong output format", "misread user intent", "ask for clarification")
	require.NoError(t, err)

	// Verify the learning was saved with category=user_correction.
	learnings, searchErr := store.SearchLearnings(ctx, "wrong output format", string(entlearning.CategoryUserCorrection), 10)
	require.NoError(t, searchErr)
	require.NotEmpty(t, learnings, "expected at least one learning after RecordUserCorrection")

	found := false
	for _, l := range learnings {
		if l.Trigger == "wrong output format" && l.Category == entlearning.CategoryUserCorrection && l.Fix == "ask for clarification" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected learning with trigger=%q, category=%q, fix=%q",
		"wrong output format", "user_correction", "ask for clarification")
}
