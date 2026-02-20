package knowledge

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent/enttest"
	_ "github.com/mattn/go-sqlite3"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	return NewStore(client, logger)
}

func TestSaveAndGetKnowledge(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("basic create and get", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:      "go-style",
			Category: "rule",
			Content:  "Use gofmt for formatting",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge: %v", err)
		}
		got, err := store.GetKnowledge(ctx, "go-style")
		if err != nil {
			t.Fatalf("GetKnowledge: %v", err)
		}
		if got.Key != "go-style" {
			t.Errorf("want key %q, got %q", "go-style", got.Key)
		}
		if got.Category != "rule" {
			t.Errorf("want category %q, got %q", "rule", got.Category)
		}
		if got.Content != "Use gofmt for formatting" {
			t.Errorf("want content %q, got %q", "Use gofmt for formatting", got.Content)
		}
	})

	t.Run("upsert overwrites content", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:      "upsert-key",
			Category: "fact",
			Content:  "original content",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge (create): %v", err)
		}
		entry.Content = "updated content"
		entry.Category = "definition"
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge (upsert): %v", err)
		}
		got, err := store.GetKnowledge(ctx, "upsert-key")
		if err != nil {
			t.Fatalf("GetKnowledge: %v", err)
		}
		if got.Content != "updated content" {
			t.Errorf("want content %q, got %q", "updated content", got.Content)
		}
		if got.Category != "definition" {
			t.Errorf("want category %q, got %q", "definition", got.Category)
		}
	})

	t.Run("optional fields tags and source", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:      "with-optional",
			Category: "preference",
			Content:  "Prefer short variable names",
			Tags:     []string{"style", "go"},
			Source:   "team-wiki",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge: %v", err)
		}
		got, err := store.GetKnowledge(ctx, "with-optional")
		if err != nil {
			t.Fatalf("GetKnowledge: %v", err)
		}
		if len(got.Tags) != 2 {
			t.Fatalf("want 2 tags, got %d", len(got.Tags))
		}
		if got.Tags[0] != "style" || got.Tags[1] != "go" {
			t.Errorf("want tags [style go], got %v", got.Tags)
		}
		if got.Source != "team-wiki" {
			t.Errorf("want source %q, got %q", "team-wiki", got.Source)
		}
	})
}

func TestGetKnowledge_NotFound(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.GetKnowledge(ctx, "nonexistent-key")
	if err == nil {
		t.Fatal("expected error for non-existent key, got nil")
	}
	if !strings.Contains(err.Error(), "knowledge not found") {
		t.Errorf("want error containing %q, got %q", "knowledge not found", err.Error())
	}
}

func TestSearchKnowledge(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entries := []KnowledgeEntry{
		{Key: "go-error-handling", Category: "rule", Content: "Always handle errors in Go"},
		{Key: "go-naming", Category: "rule", Content: "Use MixedCaps for naming"},
		{Key: "python-style", Category: "preference", Content: "Use snake_case in Python"},
	}
	for _, e := range entries {
		if err := store.SaveKnowledge(ctx, "session-1", e); err != nil {
			t.Fatalf("SaveKnowledge(%q): %v", e.Key, err)
		}
	}

	t.Run("search by content keyword", func(t *testing.T) {
		results, err := store.SearchKnowledge(ctx, "error", "", 0)
		if err != nil {
			t.Fatalf("SearchKnowledge: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
		found := false
		for _, r := range results {
			if r.Key == "go-error-handling" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find go-error-handling in results")
		}
	})

	t.Run("search by key keyword", func(t *testing.T) {
		results, err := store.SearchKnowledge(ctx, "naming", "", 0)
		if err != nil {
			t.Fatalf("SearchKnowledge: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
		if results[0].Key != "go-naming" {
			t.Errorf("want key %q, got %q", "go-naming", results[0].Key)
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		results, err := store.SearchKnowledge(ctx, "", "preference", 0)
		if err != nil {
			t.Fatalf("SearchKnowledge: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("want 1 result, got %d", len(results))
		}
		if results[0].Key != "python-style" {
			t.Errorf("want key %q, got %q", "python-style", results[0].Key)
		}
	})

	t.Run("default limit is 10", func(t *testing.T) {
		results, err := store.SearchKnowledge(ctx, "", "", 0)
		if err != nil {
			t.Fatalf("SearchKnowledge: %v", err)
		}
		if len(results) > 10 {
			t.Errorf("default limit should be 10, got %d results", len(results))
		}
	})

	t.Run("limit 1", func(t *testing.T) {
		results, err := store.SearchKnowledge(ctx, "", "", 1)
		if err != nil {
			t.Fatalf("SearchKnowledge: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("want 1 result with limit=1, got %d", len(results))
		}
	})
}

func TestDeleteKnowledge(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("delete existing entry", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:      "to-delete",
			Category: "fact",
			Content:  "Temporary fact",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge: %v", err)
		}
		if err := store.DeleteKnowledge(ctx, "to-delete"); err != nil {
			t.Fatalf("DeleteKnowledge: %v", err)
		}
		_, err := store.GetKnowledge(ctx, "to-delete")
		if err == nil {
			t.Fatal("expected error after delete, got nil")
		}
		if !strings.Contains(err.Error(), "knowledge not found") {
			t.Errorf("want error containing %q, got %q", "knowledge not found", err.Error())
		}
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		err := store.DeleteKnowledge(ctx, "no-such-key")
		if err == nil {
			t.Fatal("expected error deleting non-existent key, got nil")
		}
		if !strings.Contains(err.Error(), "knowledge not found") {
			t.Errorf("want error containing %q, got %q", "knowledge not found", err.Error())
		}
	})
}

func TestIncrementKnowledgeUseCount(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("increment existing", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:      "counter-key",
			Category: "fact",
			Content:  "Some fact",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge: %v", err)
		}
		if err := store.IncrementKnowledgeUseCount(ctx, "counter-key"); err != nil {
			t.Fatalf("IncrementKnowledgeUseCount: %v", err)
		}
		if err := store.IncrementKnowledgeUseCount(ctx, "counter-key"); err != nil {
			t.Fatalf("IncrementKnowledgeUseCount (2nd): %v", err)
		}
	})

	t.Run("non-existent key", func(t *testing.T) {
		err := store.IncrementKnowledgeUseCount(ctx, "no-such-key")
		if err == nil {
			t.Fatal("expected error for non-existent key, got nil")
		}
		if !strings.Contains(err.Error(), "knowledge not found") {
			t.Errorf("want error containing %q, got %q", "knowledge not found", err.Error())
		}
	})
}

func TestSaveLearning(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("full entry", func(t *testing.T) {
		entry := LearningEntry{
			Trigger:      "connection refused",
			ErrorPattern: "dial tcp.*connection refused",
			Diagnosis:    "Service is not running",
			Fix:          "Start the service with systemctl start",
			Category:     "tool_error",
			Tags:         []string{"network", "service"},
		}
		if err := store.SaveLearning(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveLearning: %v", err)
		}
		results, err := store.SearchLearnings(ctx, "connection refused", "", 0)
		if err != nil {
			t.Fatalf("SearchLearnings: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one learning result")
		}
		got := results[0]
		if got.Trigger != "connection refused" {
			t.Errorf("want trigger %q, got %q", "connection refused", got.Trigger)
		}
		if got.ErrorPattern != "dial tcp.*connection refused" {
			t.Errorf("want error_pattern %q, got %q", "dial tcp.*connection refused", got.ErrorPattern)
		}
		if got.Fix != "Start the service with systemctl start" {
			t.Errorf("want fix %q, got %q", "Start the service with systemctl start", got.Fix)
		}
		if got.Category != "tool_error" {
			t.Errorf("want category %q, got %q", "tool_error", got.Category)
		}
	})

	t.Run("minimal entry", func(t *testing.T) {
		entry := LearningEntry{
			Trigger:  "timeout occurred",
			Category: "timeout",
		}
		if err := store.SaveLearning(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveLearning: %v", err)
		}
		results, err := store.SearchLearnings(ctx, "timeout occurred", "", 0)
		if err != nil {
			t.Fatalf("SearchLearnings: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one learning result")
		}
		got := results[0]
		if got.Trigger != "timeout occurred" {
			t.Errorf("want trigger %q, got %q", "timeout occurred", got.Trigger)
		}
		if got.ErrorPattern != "" {
			t.Errorf("want empty error_pattern, got %q", got.ErrorPattern)
		}
	})
}

func TestSearchLearnings(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entries := []LearningEntry{
		{Trigger: "file not found", ErrorPattern: "ENOENT", Category: "tool_error"},
		{Trigger: "permission denied", ErrorPattern: "EACCES", Category: "permission"},
		{Trigger: "api timeout", ErrorPattern: "context deadline exceeded", Category: "timeout"},
	}
	for _, e := range entries {
		if err := store.SaveLearning(ctx, "session-1", e); err != nil {
			t.Fatalf("SaveLearning(%q): %v", e.Trigger, err)
		}
	}

	t.Run("filter by errorPattern", func(t *testing.T) {
		results, err := store.SearchLearnings(ctx, "ENOENT", "", 0)
		if err != nil {
			t.Fatalf("SearchLearnings: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("want 1 result, got %d", len(results))
		}
		if results[0].Trigger != "file not found" {
			t.Errorf("want trigger %q, got %q", "file not found", results[0].Trigger)
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		results, err := store.SearchLearnings(ctx, "", "permission", 0)
		if err != nil {
			t.Fatalf("SearchLearnings: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("want 1 result, got %d", len(results))
		}
		if results[0].Trigger != "permission denied" {
			t.Errorf("want trigger %q, got %q", "permission denied", results[0].Trigger)
		}
	})

	t.Run("default limit is 10", func(t *testing.T) {
		results, err := store.SearchLearnings(ctx, "", "", 0)
		if err != nil {
			t.Fatalf("SearchLearnings: %v", err)
		}
		if len(results) > 10 {
			t.Errorf("default limit should be 10, got %d results", len(results))
		}
	})
}

func TestSearchLearningEntities(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entries := []LearningEntry{
		{Trigger: "entity-trigger-1", ErrorPattern: "entity-error-1", Category: "general"},
		{Trigger: "entity-trigger-2", ErrorPattern: "entity-error-2", Category: "general"},
		{Trigger: "entity-trigger-3", ErrorPattern: "entity-error-3", Category: "general"},
	}
	for _, e := range entries {
		if err := store.SaveLearning(ctx, "session-1", e); err != nil {
			t.Fatalf("SaveLearning(%q): %v", e.Trigger, err)
		}
	}

	t.Run("returns ent entities", func(t *testing.T) {
		entities, err := store.SearchLearningEntities(ctx, "entity-error", 0)
		if err != nil {
			t.Fatalf("SearchLearningEntities: %v", err)
		}
		if len(entities) == 0 {
			t.Fatal("expected at least one entity")
		}
		// Verify they are raw ent entities with ID field
		for _, e := range entities {
			if e.ID.String() == "" {
				t.Error("expected non-empty UUID for ent entity")
			}
		}
	})

	t.Run("default limit is 5", func(t *testing.T) {
		entities, err := store.SearchLearningEntities(ctx, "entity", 0)
		if err != nil {
			t.Fatalf("SearchLearningEntities: %v", err)
		}
		if len(entities) > 5 {
			t.Errorf("default limit should be 5, got %d", len(entities))
		}
	})
}

func TestBoostLearningConfidence(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("boost increases success and occurrence", func(t *testing.T) {
		entry := LearningEntry{
			Trigger:      "boost-trigger",
			ErrorPattern: "boost-error",
			Category:     "general",
		}
		if err := store.SaveLearning(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveLearning: %v", err)
		}

		entities, err := store.SearchLearningEntities(ctx, "boost-error", 1)
		if err != nil {
			t.Fatalf("SearchLearningEntities: %v", err)
		}
		if len(entities) == 0 {
			t.Fatal("expected at least one entity")
		}

		id := entities[0].ID
		initialSuccess := entities[0].SuccessCount
		initialOccurrence := entities[0].OccurrenceCount

		if err := store.BoostLearningConfidence(ctx, id, 1, 0.0); err != nil {
			t.Fatalf("BoostLearningConfidence: %v", err)
		}

		boosted, err := store.client.Learning.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get learning: %v", err)
		}
		if boosted.SuccessCount != initialSuccess+1 {
			t.Errorf("want success_count %d, got %d", initialSuccess+1, boosted.SuccessCount)
		}
		if boosted.OccurrenceCount != initialOccurrence+1 {
			t.Errorf("want occurrence_count %d, got %d", initialOccurrence+1, boosted.OccurrenceCount)
		}
	})

	t.Run("minimum confidence is 0.1", func(t *testing.T) {
		entry := LearningEntry{
			Trigger:      "min-conf-trigger",
			ErrorPattern: "min-conf-error",
			Category:     "general",
		}
		if err := store.SaveLearning(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveLearning: %v", err)
		}

		entities, err := store.SearchLearningEntities(ctx, "min-conf-error", 1)
		if err != nil {
			t.Fatalf("SearchLearningEntities: %v", err)
		}
		if len(entities) == 0 {
			t.Fatal("expected at least one entity")
		}

		id := entities[0].ID

		// Boost with negative success delta to push confidence down
		if err := store.BoostLearningConfidence(ctx, id, -100, 0.0); err != nil {
			t.Fatalf("BoostLearningConfidence: %v", err)
		}

		boosted, err := store.client.Learning.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get learning: %v", err)
		}
		if boosted.Confidence < 0.1 {
			t.Errorf("confidence should not go below 0.1, got %f", boosted.Confidence)
		}
	})
}

func TestSaveAuditLog(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("save with all fields", func(t *testing.T) {
		entry := AuditEntry{
			SessionKey: "audit-session-1",
			Action:     "tool_call",
			Actor:      "agent-main",
			Target:     "filesystem.read",
			Details: map[string]interface{}{
				"path":   "/tmp/test.txt",
				"result": "success",
			},
		}
		if err := store.SaveAuditLog(ctx, entry); err != nil {
			t.Fatalf("SaveAuditLog: %v", err)
		}
	})

	t.Run("save with optional nil fields", func(t *testing.T) {
		entry := AuditEntry{
			Action: "knowledge_save",
			Actor:  "agent-secondary",
		}
		if err := store.SaveAuditLog(ctx, entry); err != nil {
			t.Fatalf("SaveAuditLog: %v", err)
		}
	})
}

func TestSaveAndSearchExternalRef(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("create and search by name", func(t *testing.T) {
		if err := store.SaveExternalRef(ctx, "golang-docs", "url", "https://go.dev/doc", "Official Go documentation"); err != nil {
			t.Fatalf("SaveExternalRef: %v", err)
		}
		results, err := store.SearchExternalRefs(ctx, "golang")
		if err != nil {
			t.Fatalf("SearchExternalRefs: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
		if results[0].Name != "golang-docs" {
			t.Errorf("want name %q, got %q", "golang-docs", results[0].Name)
		}
		if results[0].RefType != "url" {
			t.Errorf("want ref_type %q, got %q", "url", results[0].RefType)
		}
		if results[0].Location != "https://go.dev/doc" {
			t.Errorf("want location %q, got %q", "https://go.dev/doc", results[0].Location)
		}
	})

	t.Run("upsert updates location", func(t *testing.T) {
		if err := store.SaveExternalRef(ctx, "golang-docs", "url", "https://go.dev/doc/v2", "Updated docs"); err != nil {
			t.Fatalf("SaveExternalRef (upsert): %v", err)
		}
		results, err := store.SearchExternalRefs(ctx, "golang")
		if err != nil {
			t.Fatalf("SearchExternalRefs: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("want 1 result after upsert, got %d", len(results))
		}
		if results[0].Location != "https://go.dev/doc/v2" {
			t.Errorf("want updated location %q, got %q", "https://go.dev/doc/v2", results[0].Location)
		}
	})

	t.Run("search by summary", func(t *testing.T) {
		if err := store.SaveExternalRef(ctx, "ent-framework", "url", "https://entgo.io", "Entity framework for Go"); err != nil {
			t.Fatalf("SaveExternalRef: %v", err)
		}
		results, err := store.SearchExternalRefs(ctx, "Entity framework")
		if err != nil {
			t.Fatalf("SearchExternalRefs: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
		found := false
		for _, r := range results {
			if r.Name == "ent-framework" {
				found = true
			}
		}
		if !found {
			t.Error("expected ent-framework in search results")
		}
	})
}

func TestGetLearningStats(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("empty store", func(t *testing.T) {
		stats, err := store.GetLearningStats(ctx)
		if err != nil {
			t.Fatalf("GetLearningStats: %v", err)
		}
		if stats.TotalCount != 0 {
			t.Errorf("TotalCount: want 0, got %d", stats.TotalCount)
		}
	})

	t.Run("with entries", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			entry := LearningEntry{
				Trigger:  fmt.Sprintf("stats-trigger-%d", i),
				Category: "tool_error",
			}
			if err := store.SaveLearning(ctx, "sess-stats", entry); err != nil {
				t.Fatalf("SaveLearning %d: %v", i, err)
			}
		}
		entry := LearningEntry{
			Trigger:  "stats-trigger-general",
			Category: "general",
		}
		if err := store.SaveLearning(ctx, "sess-stats", entry); err != nil {
			t.Fatalf("SaveLearning: %v", err)
		}

		stats, err := store.GetLearningStats(ctx)
		if err != nil {
			t.Fatalf("GetLearningStats: %v", err)
		}
		if stats.TotalCount < 4 {
			t.Errorf("TotalCount: want >= 4, got %d", stats.TotalCount)
		}
		if stats.ByCategory["tool_error"] < 3 {
			t.Errorf("ByCategory[tool_error]: want >= 3, got %d", stats.ByCategory["tool_error"])
		}
	})
}

func TestDeleteLearning(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entry := LearningEntry{
		Trigger:  "delete-me",
		Category: "general",
	}
	if err := store.SaveLearning(ctx, "sess-del", entry); err != nil {
		t.Fatalf("SaveLearning: %v", err)
	}

	entities, err := store.SearchLearningEntities(ctx, "delete-me", 5)
	if err != nil || len(entities) == 0 {
		t.Fatalf("expected at least one entity, err=%v", err)
	}

	if err := store.DeleteLearning(ctx, entities[0].ID); err != nil {
		t.Fatalf("DeleteLearning: %v", err)
	}

	after, err := store.SearchLearningEntities(ctx, "delete-me", 5)
	if err != nil {
		t.Fatalf("SearchLearningEntities: %v", err)
	}
	if len(after) != 0 {
		t.Errorf("expected 0 after delete, got %d", len(after))
	}
}

func TestDeleteLearningsWhere(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		entry := LearningEntry{
			Trigger:  fmt.Sprintf("bulk-del-%d", i),
			Category: "general",
		}
		if err := store.SaveLearning(ctx, "sess-bulk", entry); err != nil {
			t.Fatalf("SaveLearning %d: %v", i, err)
		}
	}

	n, err := store.DeleteLearningsWhere(ctx, "general", 0, time.Time{})
	if err != nil {
		t.Fatalf("DeleteLearningsWhere: %v", err)
	}
	if n < 5 {
		t.Errorf("want >= 5 deleted, got %d", n)
	}
}

func TestListLearnings(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		entry := LearningEntry{
			Trigger:  fmt.Sprintf("list-test-%d", i),
			Category: "tool_error",
		}
		if err := store.SaveLearning(ctx, "sess-list", entry); err != nil {
			t.Fatalf("SaveLearning %d: %v", i, err)
		}
	}

	entries, total, err := store.ListLearnings(ctx, "tool_error", 0, time.Time{}, 2, 0)
	if err != nil {
		t.Fatalf("ListLearnings: %v", err)
	}
	if total < 3 {
		t.Errorf("total: want >= 3, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("entries: want 2 (limit), got %d", len(entries))
	}
}

