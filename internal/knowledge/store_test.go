package knowledge

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/security"
	_ "github.com/mattn/go-sqlite3"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	return NewStore(client, logger)
}

type stubPayloadProtector struct{}

func (stubPayloadProtector) EncryptPayload(plaintext []byte) ([]byte, []byte, int, error) {
	return append([]byte("enc:"), plaintext...), []byte("123456789012"), security.PayloadKeyVersionV1, nil
}

func (stubPayloadProtector) DecryptPayload(ciphertext []byte, nonce []byte, keyVersion int) ([]byte, error) {
	if keyVersion != security.PayloadKeyVersionV1 {
		return nil, fmt.Errorf("unexpected key version %d", keyVersion)
	}
	if string(nonce) != "123456789012" {
		return nil, fmt.Errorf("unexpected nonce %q", string(nonce))
	}
	return []byte(strings.TrimPrefix(string(ciphertext), "enc:")), nil
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

	t.Run("append creates new version", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:      "upsert-key",
			Category: "fact",
			Content:  "original content",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge (v1): %v", err)
		}
		entry.Content = "updated content"
		entry.Category = "definition"
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge (v2): %v", err)
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
		if got.Version != 2 {
			t.Errorf("want version 2, got %d", got.Version)
		}
		// Verify 2 DB rows exist.
		all, err := store.client.Knowledge.Query().All(ctx)
		if err != nil {
			t.Fatalf("query all: %v", err)
		}
		upsertRows := 0
		for _, k := range all {
			if k.Key == "upsert-key" {
				upsertRows++
			}
		}
		if upsertRows != 2 {
			t.Errorf("want 2 DB rows for upsert-key, got %d", upsertRows)
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

	t.Run("source tagging round trip", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:         "tagged-knowledge",
			Category:    "fact",
			Content:     "Persist classification metadata",
			Source:      "tool:save_knowledge",
			SourceClass: "user-exportable",
			AssetLabel:  "knowledge/note",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge: %v", err)
		}
		got, err := store.GetKnowledge(ctx, "tagged-knowledge")
		if err != nil {
			t.Fatalf("GetKnowledge: %v", err)
		}
		if got.SourceClass != "user-exportable" {
			t.Errorf("want source class %q, got %q", "user-exportable", got.SourceClass)
		}
		if got.AssetLabel != "knowledge/note" {
			t.Errorf("want asset label %q, got %q", "knowledge/note", got.AssetLabel)
		}
	})

	t.Run("classification change bumps version", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:         "classification-versioning",
			Category:    "fact",
			Content:     "Source metadata affects exportability",
			Source:      "tool:save_knowledge",
			SourceClass: "private-confidential",
			AssetLabel:  "knowledge/draft",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge (v1): %v", err)
		}
		entry.SourceClass = "user-exportable"
		entry.AssetLabel = "knowledge/final"
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge (v2): %v", err)
		}
		got, err := store.GetKnowledge(ctx, "classification-versioning")
		if err != nil {
			t.Fatalf("GetKnowledge: %v", err)
		}
		if got.Version != 2 {
			t.Fatalf("want version 2, got %d", got.Version)
		}
	})
}

func TestKnowledgePayloadProtection_RoundTripAndProjection(t *testing.T) {
	store := newTestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	entry := KnowledgeEntry{
		Key:      "sensitive",
		Category: "fact",
		Content:  "email alice@example.com token SECRETSECRETSECRETSECRETSECRETSECRET",
	}
	if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
		t.Fatalf("SaveKnowledge: %v", err)
	}

	got, err := store.GetKnowledge(ctx, "sensitive")
	if err != nil {
		t.Fatalf("GetKnowledge: %v", err)
	}
	if got.Content != entry.Content {
		t.Fatalf("want decrypted content %q, got %q", entry.Content, got.Content)
	}

	row, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("sensitive"), entknowledge.IsLatest(true)).
		Only(ctx)
	if err != nil {
		t.Fatalf("query raw row: %v", err)
	}
	if row.ContentCiphertext == nil || row.ContentNonce == nil || row.ContentKeyVersion == nil {
		t.Fatal("expected encrypted payload columns to be set")
	}
	if *row.ContentKeyVersion != security.PayloadKeyVersionV1 {
		t.Fatalf("want key version %d, got %d", security.PayloadKeyVersionV1, *row.ContentKeyVersion)
	}
	if strings.Contains(row.Content, "alice@example.com") {
		t.Fatalf("projection leaked email: %q", row.Content)
	}
	if strings.Contains(row.Content, "SECRETSECRETSECRETSECRETSECRETSECRET") {
		t.Fatalf("projection leaked secret token: %q", row.Content)
	}
	if !strings.Contains(row.Content, "[email]") || !strings.Contains(row.Content, "[secret]") {
		t.Fatalf("projection did not redact sensitive material: %q", row.Content)
	}
}

func TestLearningPayloadProtection_RoundTripAndProjection(t *testing.T) {
	store := newTestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	entry := LearningEntry{
		Trigger:      "timeout",
		ErrorPattern: "email alice@example.com",
		Diagnosis:    "token SECRETSECRETSECRETSECRETSECRETSECRET",
		Fix:          "rotate 123456789",
		Category:     entlearning.CategoryTimeout,
	}
	require.NoError(t, store.SaveLearning(ctx, "session-1", entry))

	got, err := store.GetLearning(ctx, store.client.Learning.Query().OnlyX(ctx).ID)
	require.NoError(t, err)
	require.Equal(t, entry.ErrorPattern, got.ErrorPattern)
	require.Equal(t, entry.Diagnosis, got.Diagnosis)
	require.Equal(t, entry.Fix, got.Fix)

	results, err := store.SearchLearningsScored(ctx, "timeout", "", 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, entry.ErrorPattern, results[0].Entry.ErrorPattern)

	row := store.client.Learning.Query().OnlyX(ctx)
	require.NotNil(t, row.PayloadCiphertext)
	require.NotContains(t, row.ErrorPattern, "alice@example.com")
	require.NotContains(t, row.Diagnosis, "SECRETSECRETSECRETSECRETSECRETSECRET")
	require.NotContains(t, row.Fix, "123456789")
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

	t.Run("delete all versions", func(t *testing.T) {
		entry := KnowledgeEntry{
			Key:      "to-delete",
			Category: "fact",
			Content:  "v1 content",
		}
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge (v1): %v", err)
		}
		entry.Content = "v2 content"
		if err := store.SaveKnowledge(ctx, "session-1", entry); err != nil {
			t.Fatalf("SaveKnowledge (v2): %v", err)
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
		// Verify no rows remain for this key.
		_, err = store.GetKnowledgeHistory(ctx, "to-delete")
		if err == nil {
			t.Fatal("expected history error after delete, got nil")
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
			Category:     entlearning.CategoryToolError,
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
		if got.Category != entlearning.CategoryToolError {
			t.Errorf("want category %q, got %q", entlearning.CategoryToolError, got.Category)
		}
	})

	t.Run("minimal entry", func(t *testing.T) {
		entry := LearningEntry{
			Trigger:  "timeout occurred",
			Category: entlearning.CategoryTimeout,
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
		{Trigger: "file not found", ErrorPattern: "ENOENT", Category: entlearning.CategoryToolError},
		{Trigger: "permission denied", ErrorPattern: "EACCES", Category: entlearning.CategoryPermission},
		{Trigger: "api timeout", ErrorPattern: "context deadline exceeded", Category: entlearning.CategoryTimeout},
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
		{Trigger: "entity-trigger-1", ErrorPattern: "entity-error-1", Category: entlearning.CategoryGeneral},
		{Trigger: "entity-trigger-2", ErrorPattern: "entity-error-2", Category: entlearning.CategoryGeneral},
		{Trigger: "entity-trigger-3", ErrorPattern: "entity-error-3", Category: entlearning.CategoryGeneral},
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
			Category:     entlearning.CategoryGeneral,
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
			Category:     entlearning.CategoryGeneral,
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
				Category: entlearning.CategoryToolError,
			}
			if err := store.SaveLearning(ctx, "sess-stats", entry); err != nil {
				t.Fatalf("SaveLearning %d: %v", i, err)
			}
		}
		entry := LearningEntry{
			Trigger:  "stats-trigger-general",
			Category: entlearning.CategoryGeneral,
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
		if stats.ByCategory[entlearning.CategoryToolError] < 3 {
			t.Errorf("ByCategory[tool_error]: want >= 3, got %d", stats.ByCategory[entlearning.CategoryToolError])
		}
	})
}

func TestDeleteLearning(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entry := LearningEntry{
		Trigger:  "delete-me",
		Category: entlearning.CategoryGeneral,
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
			Category: entlearning.CategoryGeneral,
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
			Category: entlearning.CategoryToolError,
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

// ── EventBus graph routing regression tests ──
// These verify that NeedsGraph is correctly set per save path,
// preserving the original SetGraphCallback behavior.

func TestContentSavedEvent_NeedsGraph(t *testing.T) {
	store := newTestStore(t)
	bus := eventbus.New()
	store.SetEventBus(bus)
	ctx := context.Background()

	var events []eventbus.ContentSavedEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.ContentSavedEvent) {
		events = append(events, evt)
	})

	tests := []struct {
		give       string
		wantGraph  bool
		wantNew    bool
		collection string
		setup      func()
	}{
		{
			give:       "new knowledge creation triggers graph",
			wantGraph:  true,
			wantNew:    true,
			collection: "knowledge",
			setup: func() {
				events = nil
				_ = store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
					Key: "test-new", Category: "fact", Content: "new content",
				})
			},
		},
		{
			give:       "knowledge update skips graph",
			wantGraph:  false,
			wantNew:    false,
			collection: "knowledge",
			setup: func() {
				events = nil
				_ = store.SaveKnowledge(ctx, "s1", KnowledgeEntry{
					Key: "test-new", Category: "fact", Content: "updated content",
				})
			},
		},
		{
			give:       "learning save skips graph",
			wantGraph:  false,
			wantNew:    true,
			collection: "learning",
			setup: func() {
				events = nil
				_ = store.SaveLearning(ctx, "s1", LearningEntry{
					Category: entlearning.CategoryToolError,
					Trigger:  "test trigger",
					Fix:      "test fix",
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tt.setup()
			if len(events) != 1 {
				t.Fatalf("want 1 event, got %d", len(events))
			}
			evt := events[0]
			if evt.NeedsGraph != tt.wantGraph {
				t.Errorf("NeedsGraph: want %v, got %v", tt.wantGraph, evt.NeedsGraph)
			}
			if evt.IsNew != tt.wantNew {
				t.Errorf("IsNew: want %v, got %v", tt.wantNew, evt.IsNew)
			}
			if evt.Collection != tt.collection {
				t.Errorf("Collection: want %q, got %q", tt.collection, evt.Collection)
			}
		})
	}
}

func TestSaveKnowledge_VersionHistory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		entry := KnowledgeEntry{
			Key:      "versioned-key",
			Category: "fact",
			Content:  fmt.Sprintf("content v%d", i),
		}
		if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
			t.Fatalf("SaveKnowledge v%d: %v", i, err)
		}
	}

	got, err := store.GetKnowledge(ctx, "versioned-key")
	if err != nil {
		t.Fatalf("GetKnowledge: %v", err)
	}
	if got.Version != 3 {
		t.Errorf("want version 3, got %d", got.Version)
	}
	if got.Content != "content v3" {
		t.Errorf("want content %q, got %q", "content v3", got.Content)
	}

	// Verify 3 DB rows exist.
	all, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("versioned-key")).
		All(ctx)
	if err != nil {
		t.Fatalf("query all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("want 3 rows, got %d", len(all))
	}

	// Exactly one is_latest=true.
	latestCount := 0
	for _, k := range all {
		if k.IsLatest {
			latestCount++
		}
	}
	if latestCount != 1 {
		t.Errorf("want exactly 1 is_latest=true, got %d", latestCount)
	}
}

func TestGetKnowledgeHistory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		entry := KnowledgeEntry{
			Key:      "hist-key",
			Category: "fact",
			Content:  fmt.Sprintf("content v%d", i),
		}
		if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
			t.Fatalf("SaveKnowledge v%d: %v", i, err)
		}
	}

	history, err := store.GetKnowledgeHistory(ctx, "hist-key")
	if err != nil {
		t.Fatalf("GetKnowledgeHistory: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("want 3 versions, got %d", len(history))
	}
	// Descending order.
	if history[0].Version != 3 || history[1].Version != 2 || history[2].Version != 1 {
		t.Errorf("want versions [3,2,1], got [%d,%d,%d]", history[0].Version, history[1].Version, history[2].Version)
	}
	// CreatedAt populated.
	for _, h := range history {
		if h.CreatedAt.IsZero() {
			t.Errorf("version %d: CreatedAt should not be zero", h.Version)
		}
	}
}

func TestGetKnowledgeHistory_NotFound(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.GetKnowledgeHistory(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "knowledge not found") {
		t.Errorf("want error containing %q, got %q", "knowledge not found", err.Error())
	}
}

func TestSaveKnowledge_CarryForward(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entry := KnowledgeEntry{
		Key:      "carry-key",
		Category: "fact",
		Content:  "v1 content",
	}
	if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
		t.Fatalf("SaveKnowledge v1: %v", err)
	}

	// Increment use count on v1.
	if err := store.IncrementKnowledgeUseCount(ctx, "carry-key"); err != nil {
		t.Fatalf("IncrementKnowledgeUseCount: %v", err)
	}
	if err := store.IncrementKnowledgeUseCount(ctx, "carry-key"); err != nil {
		t.Fatalf("IncrementKnowledgeUseCount (2nd): %v", err)
	}

	// Save v2 — should carry forward use_count=2.
	entry.Content = "v2 content"
	if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
		t.Fatalf("SaveKnowledge v2: %v", err)
	}

	// Read back v2 via direct DB query.
	latest, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("carry-key"), entknowledge.IsLatest(true)).
		Only(ctx)
	if err != nil {
		t.Fatalf("query latest: %v", err)
	}
	if latest.UseCount != 2 {
		t.Errorf("want use_count 2 (carried forward), got %d", latest.UseCount)
	}
	if latest.Version != 2 {
		t.Errorf("want version 2, got %d", latest.Version)
	}
}

func TestSearchKnowledge_LatestOnly(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Save v1 with "old unicorn data".
	entry := KnowledgeEntry{
		Key:      "search-latest",
		Category: "fact",
		Content:  "old unicorn data",
	}
	if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
		t.Fatalf("SaveKnowledge v1: %v", err)
	}

	// Save v2 with completely different content.
	entry.Content = "new dragon data"
	if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
		t.Fatalf("SaveKnowledge v2: %v", err)
	}

	// Search for "unicorn" — should find nothing (v1 is not latest).
	results, err := store.SearchKnowledge(ctx, "unicorn", "", 10)
	if err != nil {
		t.Fatalf("SearchKnowledge: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("want 0 results for 'unicorn' (old version), got %d", len(results))
	}

	// Search for "dragon" — should find latest version.
	results, err = store.SearchKnowledge(ctx, "dragon", "", 10)
	if err != nil {
		t.Fatalf("SearchKnowledge: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("want 1 result for 'dragon' (latest version), got %d", len(results))
	}
}

func TestSaveKnowledge_ContentDedup(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entry := KnowledgeEntry{
		Key:      "dedup-key",
		Category: "fact",
		Content:  "Go is a compiled language",
	}

	// First save — creates version 1.
	if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
		t.Fatalf("SaveKnowledge v1: %v", err)
	}

	// Second save with SAME category+content — should be no-op.
	if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
		t.Fatalf("SaveKnowledge dedup: %v", err)
	}

	// Verify only 1 version exists.
	history, err := store.GetKnowledgeHistory(ctx, "dedup-key")
	if err != nil {
		t.Fatalf("GetKnowledgeHistory: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("want 1 version (dedup), got %d", len(history))
	}

	// Third save with DIFFERENT content — should create version 2.
	entry.Content = "Go is a statically typed, compiled language"
	if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
		t.Fatalf("SaveKnowledge v2: %v", err)
	}

	history, err = store.GetKnowledgeHistory(ctx, "dedup-key")
	if err != nil {
		t.Fatalf("GetKnowledgeHistory: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("want 2 versions, got %d", len(history))
	}

	// Fourth save with same content but different category — should create version 3.
	entry.Category = "definition"
	if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
		t.Fatalf("SaveKnowledge v3: %v", err)
	}

	history, err = store.GetKnowledgeHistory(ctx, "dedup-key")
	if err != nil {
		t.Fatalf("GetKnowledgeHistory: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("want 3 versions, got %d", len(history))
	}
}

func TestBoostRelevanceScore_Clamping(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		give      string
		initial   float64
		delta     float64
		maxScore  float64
		wantScore float64
	}{
		{
			give:      "within range: 4.94 + 0.05 = 4.99",
			initial:   4.94,
			delta:     0.05,
			maxScore:  5.0,
			wantScore: 4.99,
		},
		{
			give:      "exact boundary: 4.95 + 0.05 = 5.00",
			initial:   4.95,
			delta:     0.05,
			maxScore:  5.0,
			wantScore: 5.0,
		},
		{
			give:      "overshoot capped: 4.96 + 0.05 -> 5.00",
			initial:   4.96,
			delta:     0.05,
			maxScore:  5.0,
			wantScore: 5.0,
		},
		{
			give:      "already at max: 5.00 + 0.05 -> 5.00",
			initial:   5.0,
			delta:     0.05,
			maxScore:  5.0,
			wantScore: 5.0,
		},
		{
			give:      "pre-existing over-cap normalized: 5.10 + 0.05 -> 5.00",
			initial:   5.10,
			delta:     0.05,
			maxScore:  5.0,
			wantScore: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			key := fmt.Sprintf("boost-%s", tt.give)
			entry := KnowledgeEntry{
				Key: key, Category: "fact", Content: "test",
			}
			if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
				t.Fatalf("SaveKnowledge: %v", err)
			}
			// Set initial relevance score.
			_, err := store.client.Knowledge.Update().
				Where(entknowledge.Key(key), entknowledge.IsLatest(true)).
				SetRelevanceScore(tt.initial).
				Save(ctx)
			if err != nil {
				t.Fatalf("set initial score: %v", err)
			}

			if err := store.BoostRelevanceScore(ctx, key, tt.delta, tt.maxScore); err != nil {
				t.Fatalf("BoostRelevanceScore: %v", err)
			}

			k, err := store.client.Knowledge.Query().
				Where(entknowledge.Key(key), entknowledge.IsLatest(true)).
				Only(ctx)
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if k.RelevanceScore != tt.wantScore {
				t.Errorf("want score %.2f, got %.2f", tt.wantScore, k.RelevanceScore)
			}
		})
	}
}

func TestDecayAllRelevanceScores_Clamping(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		give      string
		initial   float64
		delta     float64
		minScore  float64
		wantScore float64
	}{
		{
			give:      "within range: 0.16 - 0.05 = 0.11",
			initial:   0.16,
			delta:     0.05,
			minScore:  0.1,
			wantScore: 0.11,
		},
		{
			give:      "exact boundary: 0.15 - 0.05 = 0.10",
			initial:   0.15,
			delta:     0.05,
			minScore:  0.1,
			wantScore: 0.1,
		},
		{
			give:      "undershoot floored: 0.14 - 0.05 -> 0.10",
			initial:   0.14,
			delta:     0.05,
			minScore:  0.1,
			wantScore: 0.1,
		},
		{
			give:      "already at min: 0.10 - 0.05 -> 0.10",
			initial:   0.1,
			delta:     0.05,
			minScore:  0.1,
			wantScore: 0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			key := fmt.Sprintf("decay-%s", tt.give)
			entry := KnowledgeEntry{
				Key: key, Category: "fact", Content: "test",
			}
			if err := store.SaveKnowledge(ctx, "s1", entry); err != nil {
				t.Fatalf("SaveKnowledge: %v", err)
			}
			_, err := store.client.Knowledge.Update().
				Where(entknowledge.Key(key), entknowledge.IsLatest(true)).
				SetRelevanceScore(tt.initial).
				Save(ctx)
			if err != nil {
				t.Fatalf("set initial score: %v", err)
			}

			_, err = store.DecayAllRelevanceScores(ctx, tt.delta, tt.minScore)
			if err != nil {
				t.Fatalf("DecayAllRelevanceScores: %v", err)
			}

			k, err := store.client.Knowledge.Query().
				Where(entknowledge.Key(key), entknowledge.IsLatest(true)).
				Only(ctx)
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if k.RelevanceScore != tt.wantScore {
				t.Errorf("want score %.2f, got %.2f", tt.wantScore, k.RelevanceScore)
			}
		})
	}
}
