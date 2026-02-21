package learning

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent/enttest"
	entlearning "github.com/langowarny/lango/internal/ent/learning"
	"github.com/langowarny/lango/internal/knowledge"
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
	engine, _ := newTestEngine(t)
	ctx := context.Background()

	// Calling OnToolResult with nil error should not panic and should save an audit log.
	engine.OnToolResult(ctx, "sess-1", "file_read", map[string]interface{}{"path": "/tmp"}, "ok", nil)
}

func TestEngine_OnToolResult_Error_NewPattern(t *testing.T) {
	engine, store := newTestEngine(t)
	ctx := context.Background()

	testErr := errors.New("connection refused")
	engine.OnToolResult(ctx, "sess-1", "http_call", nil, nil, testErr)

	// Verify a new learning was created by searching for the error pattern.
	learnings, err := store.SearchLearnings(ctx, "connection refused", "", 10)
	if err != nil {
		t.Fatalf("SearchLearnings: %v", err)
	}
	if len(learnings) == 0 {
		t.Fatal("expected at least one learning after OnToolResult with error, got 0")
	}

	found := false
	for _, l := range learnings {
		if l.Trigger == "tool:http_call" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected learning with trigger %q, got %v", "tool:http_call", learnings)
	}
}

func TestEngine_OnToolResult_Error_KnownFix(t *testing.T) {
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
	if err != nil {
		t.Fatalf("SaveLearning: %v", err)
	}

	// Boost confidence above 0.5 by searching and updating directly.
	entities, err := store.SearchLearningEntities(ctx, "connection refused", 5)
	if err != nil {
		t.Fatalf("SearchLearningEntities: %v", err)
	}
	if len(entities) == 0 {
		t.Fatal("expected at least one entity")
	}

	// Set confidence to 0.8 directly via ent update.
	_, err = entities[0].Update().SetConfidence(0.8).SetSuccessCount(10).Save(ctx)
	if err != nil {
		t.Fatalf("update confidence: %v", err)
	}

	// Count learnings before calling OnToolResult with a matching error.
	before, err := store.SearchLearnings(ctx, "connection refused", "", 50)
	if err != nil {
		t.Fatalf("SearchLearnings before: %v", err)
	}
	beforeCount := len(before)

	// Call OnToolResult with matching error - should NOT create a new learning
	// because a high-confidence fix already exists.
	testErr := errors.New("connection refused")
	engine.OnToolResult(ctx, "sess-2", "http_call", nil, nil, testErr)

	after, err := store.SearchLearnings(ctx, "connection refused", "", 50)
	if err != nil {
		t.Fatalf("SearchLearnings after: %v", err)
	}
	if len(after) != beforeCount {
		t.Errorf("expected no new learning (count %d), but got %d", beforeCount, len(after))
	}
}

func TestEngine_GetFixForError(t *testing.T) {
	engine, store := newTestEngine(t)
	ctx := context.Background()

	t.Run("returns fix for high-confidence learning", func(t *testing.T) {
		// Use an error message that extractErrorPattern returns unchanged
		// and that the Contains-based search can match.
		// SearchLearningEntities searches: stored_field CONTAINS query.
		// So the stored error_pattern must contain the extracted pattern from the test error.
		errMsg := "undefined variable in scope"
		err := store.SaveLearning(ctx, "sess-1", knowledge.LearningEntry{
			Trigger:      "tool:compile",
			ErrorPattern: errMsg,
			Diagnosis:    "missing declaration",
			Fix:          "declare the variable before use",
			Category:     entlearning.CategoryToolError,
		})
		if err != nil {
			t.Fatalf("SaveLearning: %v", err)
		}

		// Set confidence above 0.5.
		entities, err := store.SearchLearningEntities(ctx, errMsg, 5)
		if err != nil {
			t.Fatalf("SearchLearningEntities: %v", err)
		}
		if len(entities) == 0 {
			t.Fatal("expected at least one entity")
		}
		_, err = entities[0].Update().SetConfidence(0.8).Save(ctx)
		if err != nil {
			t.Fatalf("update confidence: %v", err)
		}

		// GetFixForError extracts pattern from error, then searches with Contains.
		// The stored error_pattern "undefined variable in scope" contains "undefined variable in scope".
		fix, ok := engine.GetFixForError(ctx, "compile", errors.New(errMsg))
		if !ok {
			t.Fatal("GetFixForError returned false, want true")
		}
		if fix != "declare the variable before use" {
			t.Errorf("fix = %q, want %q", fix, "declare the variable before use")
		}
	})

	t.Run("returns false for non-matching error", func(t *testing.T) {
		fix, ok := engine.GetFixForError(ctx, "compile", errors.New("completely unrelated xyz error"))
		if ok {
			t.Errorf("GetFixForError returned true for non-matching error, fix = %q", fix)
		}
		if fix != "" {
			t.Errorf("fix = %q, want empty string", fix)
		}
	})

	t.Run("returns false for low-confidence learning", func(t *testing.T) {
		err := store.SaveLearning(ctx, "sess-2", knowledge.LearningEntry{
			Trigger:      "tool:deploy",
			ErrorPattern: "low conf pattern xyz",
			Diagnosis:    "some diagnosis",
			Fix:          "some fix",
			Category:     entlearning.CategoryToolError,
		})
		if err != nil {
			t.Fatalf("SaveLearning: %v", err)
		}

		// Set confidence below 0.5.
		entities, err := store.SearchLearningEntities(ctx, "low conf pattern xyz", 5)
		if err != nil {
			t.Fatalf("SearchLearningEntities: %v", err)
		}
		if len(entities) == 0 {
			t.Fatal("expected at least one entity")
		}
		_, err = entities[0].Update().SetConfidence(0.3).Save(ctx)
		if err != nil {
			t.Fatalf("update confidence: %v", err)
		}

		fix, ok := engine.GetFixForError(ctx, "deploy", errors.New("low conf pattern xyz"))
		if ok {
			t.Errorf("GetFixForError returned true for low-confidence learning, fix = %q", fix)
		}
		if fix != "" {
			t.Errorf("fix = %q, want empty string", fix)
		}
	})
}

func TestEngine_RecordUserCorrection(t *testing.T) {
	engine, store := newTestEngine(t)
	ctx := context.Background()

	err := engine.RecordUserCorrection(ctx, "sess-1", "wrong output format", "misread user intent", "ask for clarification")
	if err != nil {
		t.Fatalf("RecordUserCorrection: %v", err)
	}

	// Verify the learning was saved with category=user_correction.
	learnings, searchErr := store.SearchLearnings(ctx, "wrong output format", string(entlearning.CategoryUserCorrection), 10)
	if searchErr != nil {
		t.Fatalf("SearchLearnings: %v", searchErr)
	}
	if len(learnings) == 0 {
		t.Fatal("expected at least one learning after RecordUserCorrection, got 0")
	}

	found := false
	for _, l := range learnings {
		if l.Trigger == "wrong output format" && l.Category == entlearning.CategoryUserCorrection && l.Fix == "ask for clarification" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected learning with trigger=%q, category=%q, fix=%q; got %v",
			"wrong output format", "user_correction", "ask for clarification", learnings)
	}
}
