package knowledge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/auditlog"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/security"
)

type failingWave19PayloadProtector struct{}

func (failingWave19PayloadProtector) EncryptPayload([]byte) ([]byte, []byte, int, error) {
	return nil, nil, 0, errors.New("encrypt failed")
}

func (failingWave19PayloadProtector) DecryptPayload([]byte, []byte, int) ([]byte, error) {
	return nil, errors.New("decrypt failed")
}

func TestWave19SaveKnowledgeNormalizesSourceClassDedupsAndCarriesState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entry := KnowledgeEntry{
		Key:         "wave19-version-state",
		Category:    entknowledge.CategoryFact,
		Content:     "versioned content",
		Tags:        []string{"wave19", "state"},
		Source:      "tool:first",
		SourceClass: "  public  ",
		AssetLabel:  "knowledge/wave19",
	}
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:         entry.Key,
		Category:    entry.Category,
		Content:     entry.Content,
		Tags:        []string{"ignored", "for-dedup"},
		Source:      entry.Source,
		SourceClass: "public",
		AssetLabel:  entry.AssetLabel,
	}))

	got, err := store.GetKnowledge(ctx, entry.Key)
	require.NoError(t, err)
	require.Equal(t, 1, got.Version)
	require.Equal(t, "public", got.SourceClass)
	require.Equal(t, []string{"wave19", "state"}, got.Tags)

	require.NoError(t, store.IncrementKnowledgeUseCount(ctx, entry.Key))
	require.NoError(t, store.BoostRelevanceScore(ctx, entry.Key, 1.25, 10))

	entry.Source = "tool:second"
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))

	latest, err := store.client.Knowledge.Query().
		Where(entknowledge.Key(entry.Key), entknowledge.IsLatest(true)).
		Only(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, latest.Version)
	require.Equal(t, "tool:second", latest.Source)
	require.Equal(t, 1, latest.UseCount)
	require.Equal(t, 2.25, latest.RelevanceScore)

	history, err := store.GetKnowledgeHistory(ctx, entry.Key)
	require.NoError(t, err)
	require.Len(t, history, 2)
	require.Equal(t, "tool:second", history[0].Source)
	require.Equal(t, "tool:first", history[1].Source)
}

func TestWave19KnowledgeSearchAndKeywordHelpersHandleWhitespaceAndLatestOnly(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave19-alpha",
		Category: entknowledge.CategoryFact,
		Content:  "retired token",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave19-alpha",
		Category: entknowledge.CategoryFact,
		Content:  "current alpha guidance",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave19-beta-key",
		Category: entknowledge.CategoryRule,
		Content:  "rule content",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave19-gamma",
		Category: entknowledge.CategoryFact,
		Content:  "gamma material",
	}))

	preds := knowledgeKeywordPredicates("  alpha   beta  ")
	require.Len(t, preds, 4)
	require.Empty(t, knowledgeKeywordPredicates(" \t\n "))

	got, err := store.SearchKnowledge(ctx, "retired", "", 10)
	require.NoError(t, err)
	require.Empty(t, got)

	got, err = store.SearchKnowledge(ctx, "alpha beta", "", 10)
	require.NoError(t, err)
	require.Len(t, got, 2)
	keys := []string{got[0].Key, got[1].Key}
	require.ElementsMatch(t, []string{"wave19-alpha", "wave19-beta-key"}, keys)

	facts, err := store.SearchKnowledge(ctx, "alpha beta", string(entknowledge.CategoryFact), 10)
	require.NoError(t, err)
	require.Len(t, facts, 1)
	require.Equal(t, "wave19-alpha", facts[0].Key)
}

func TestWave19LearningStatsUseExactAggregatesAndOrderingBounds(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	base := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	_, err := store.client.Learning.Create().
		SetTrigger("wave19-old-timeout").
		SetCategory(entlearning.CategoryTimeout).
		SetConfidence(0.2).
		SetOccurrenceCount(3).
		SetSuccessCount(1).
		SetCreatedAt(base.Add(-2 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.Create().
		SetTrigger("wave19-new-timeout").
		SetCategory(entlearning.CategoryTimeout).
		SetConfidence(0.8).
		SetOccurrenceCount(5).
		SetSuccessCount(4).
		SetCreatedAt(base).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.Create().
		SetTrigger("wave19-permission").
		SetCategory(entlearning.CategoryPermission).
		SetConfidence(0.5).
		SetOccurrenceCount(2).
		SetSuccessCount(1).
		SetCreatedAt(base.Add(-1 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	stats, err := store.GetLearningStats(ctx)
	require.NoError(t, err)
	require.Equal(t, 3, stats.TotalCount)
	require.Equal(t, 2, stats.ByCategory[entlearning.CategoryTimeout])
	require.Equal(t, 1, stats.ByCategory[entlearning.CategoryPermission])
	require.InDelta(t, 0.5, stats.AvgConfidence, 0.000001)
	require.Equal(t, 10, stats.TotalOccurrences)
	require.Equal(t, 6, stats.TotalSuccesses)
	require.Equal(t, base.Add(-2*time.Hour), stats.OldestEntry)
	require.Equal(t, base, stats.NewestEntry)

	listed, total, err := store.ListLearnings(ctx, "", 0, time.Time{}, 2, 1)
	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, listed, 2)
	require.Equal(t, "wave19-permission", listed[0].Trigger)
	require.Equal(t, "wave19-old-timeout", listed[1].Trigger)
}

func TestWave19LearningAndExternalRefHelpersFilterPrecisely(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	cutoff := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	_, err := store.client.Learning.Create().
		SetTrigger("wave19-delete-timeout-low-old").
		SetErrorPattern("alpha timeout").
		SetCategory(entlearning.CategoryTimeout).
		SetConfidence(0.2).
		SetCreatedAt(cutoff.Add(-2 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.Create().
		SetTrigger("wave19-keep-timeout-high-old").
		SetErrorPattern("alpha timeout high").
		SetCategory(entlearning.CategoryTimeout).
		SetConfidence(0.9).
		SetCreatedAt(cutoff.Add(-2 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.Create().
		SetTrigger("wave19-keep-permission-low-old").
		SetErrorPattern("alpha permission").
		SetCategory(entlearning.CategoryPermission).
		SetConfidence(0.2).
		SetCreatedAt(cutoff.Add(-2 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	require.Len(t, learningKeywordPredicates(" alpha   timeout "), 4)
	require.Empty(t, learningKeywordPredicates("   "))

	deleted, err := store.DeleteLearningsWhere(
		ctx,
		string(entlearning.CategoryTimeout),
		0.5,
		cutoff.Add(-time.Hour),
	)
	require.NoError(t, err)
	require.Equal(t, 1, deleted)

	remaining, err := store.SearchLearningEntities(ctx, "wave19", 10)
	require.NoError(t, err)
	triggers := make([]string, 0, len(remaining))
	for _, row := range remaining {
		triggers = append(triggers, row.Trigger)
	}
	require.ElementsMatch(t, []string{
		"wave19-keep-timeout-high-old",
		"wave19-keep-permission-low-old",
	}, triggers)

	require.Len(t, externalRefKeywordPredicates(" go   docs "), 4)
	require.Empty(t, externalRefKeywordPredicates("   "))

	require.NoError(t, store.SaveExternalRef(
		ctx,
		"wave19-reference",
		"url",
		"https://example.test/old",
		"initial summary",
	))
	require.NoError(t, store.SaveExternalRef(
		ctx,
		"wave19-reference",
		"file",
		"/tmp/wave19-reference.md",
		"updated summary",
	))

	refs, err := store.SearchExternalRefs(ctx, "updated reference")
	require.NoError(t, err)
	require.Len(t, refs, 1)
	require.Equal(t, "file", refs[0].RefType)
	require.Equal(t, "/tmp/wave19-reference.md", refs[0].Location)
	require.Equal(t, "updated summary", refs[0].Summary)
}

func TestWave19PayloadProtectionFailurePathsAreObservable(t *testing.T) {
	t.Run("knowledge write encryption error is wrapped", func(t *testing.T) {
		store := newTestStore(t)
		store.SetPayloadProtector(failingWave19PayloadProtector{})
		ctx := context.Background()

		err := store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
			Key:      "wave19-encrypt-fail",
			Category: entknowledge.CategoryFact,
			Content:  "cannot encrypt",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "prepare knowledge payload")
		require.Contains(t, err.Error(), "encrypt failed")
	})

	t.Run("knowledge read falls back to stored projection on decrypt error", func(t *testing.T) {
		store := newTestStore(t)
		store.SetPayloadProtector(failingWave19PayloadProtector{})
		ctx := context.Background()

		_, err := store.client.Knowledge.Create().
			SetKey("wave19-decrypt-fallback").
			SetCategory(entknowledge.CategoryFact).
			SetContent("redacted projection").
			SetVersion(1).
			SetIsLatest(true).
			SetContentCiphertext([]byte("ciphertext")).
			SetContentNonce([]byte("123456789012")).
			SetContentKeyVersion(security.PayloadKeyVersionV1).
			Save(ctx)
		require.NoError(t, err)

		got, err := store.GetKnowledge(ctx, "wave19-decrypt-fallback")
		require.NoError(t, err)
		require.Equal(t, "redacted projection", got.Content)
	})

	t.Run("learning write and read decryption errors are wrapped", func(t *testing.T) {
		store := newTestStore(t)
		store.SetPayloadProtector(failingWave19PayloadProtector{})
		ctx := context.Background()

		err := store.SaveLearning(ctx, "session-1", LearningEntry{
			Trigger:      "wave19 learning encrypt",
			ErrorPattern: "cannot encrypt",
			Category:     entlearning.CategoryToolError,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "prepare learning payload")
		require.Contains(t, err.Error(), "encrypt failed")

		id := uuid.New()
		_, err = store.client.Learning.Create().
			SetID(id).
			SetTrigger("wave19 learning decrypt").
			SetCategory(entlearning.CategoryToolError).
			SetPayloadCiphertext([]byte("ciphertext")).
			SetPayloadNonce([]byte("123456789012")).
			SetPayloadKeyVersion(security.PayloadKeyVersionV1).
			Save(ctx)
		require.NoError(t, err)

		_, err = store.GetLearning(ctx, id)
		require.Error(t, err)
		require.Contains(t, err.Error(), "resolve learning")
		require.Contains(t, err.Error(), "decrypt learning payload")
	})
}

func TestWave19ClosedBackendErrorsCoverStoreOperations(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	require.NoError(t, store.client.Close())

	_, err := store.GetKnowledgeByKeys(ctx, []string{"missing"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "query knowledge by keys")

	_, err = store.GetKnowledgeHistory(ctx, "missing")
	require.Error(t, err)
	require.Contains(t, err.Error(), "query knowledge history")

	_, err = store.SearchKnowledgeScored(ctx, "anything", "", 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "search knowledge scored")

	err = store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:  "closed learning",
		Category: entlearning.CategoryGeneral,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "create learning")

	_, err = store.GetLearning(ctx, uuid.New())
	require.Error(t, err)
	require.Contains(t, err.Error(), "get learning")

	_, err = store.SearchLearnings(ctx, "anything", "", 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "search learnings")

	_, err = store.SearchLearningsScored(ctx, "anything", "", 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "search learnings scored")

	_, err = store.SearchLearningEntities(ctx, "anything", 10)
	require.Error(t, err)

	err = store.BoostLearningConfidence(ctx, uuid.New(), 1, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "get learning")

	err = store.SaveAuditLog(ctx, AuditEntry{
		Action: string(auditlog.ActionKnowledgeSave),
		Actor:  "wave19",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "create audit log")

	err = store.SaveExternalRef(ctx, "closed-ref", "url", "https://example.test", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "query external ref")

	_, err = store.SearchExternalRefs(ctx, "anything")
	require.Error(t, err)
	require.Contains(t, err.Error(), "search external refs")

	_, err = store.GetLearningStats(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "query learnings")

	_, _, err = store.ListLearnings(ctx, "", 0, time.Time{}, 10, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "count learnings")

	err = store.DeleteLearning(ctx, uuid.New())
	require.Error(t, err)
	require.Contains(t, err.Error(), "delete learning")

	_, err = store.DeleteLearningsWhere(ctx, string(entlearning.CategoryGeneral), 0, time.Time{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "delete learnings")
}

func TestWave19SaveAuditLogPersistsOptionalFields(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.SaveAuditLog(ctx, AuditEntry{
		SessionKey: "wave19-session",
		Action:     string(auditlog.ActionToolCall),
		Actor:      "wave19-agent",
		Target:     "knowledge.store",
		Details: map[string]interface{}{
			"result": "ok",
			"count":  float64(2),
		},
	})
	require.NoError(t, err)

	row, err := store.client.AuditLog.Query().
		Where(auditlog.SessionKey("wave19-session")).
		Only(ctx)
	require.NoError(t, err)
	require.Equal(t, auditlog.ActionToolCall, row.Action)
	require.Equal(t, "wave19-agent", row.Actor)
	require.Equal(t, "knowledge.store", row.Target)
	require.Equal(t, "ok", row.Details["result"])
	require.Equal(t, float64(2), row.Details["count"])
}

func TestWave19SaveToolResultFormatsNonStringResult(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.SaveToolResult(
		ctx,
		"session-1",
		"calculator",
		nil,
		struct {
			Value int
			Unit  string
		}{Value: 7, Unit: "items"},
	)
	require.NoError(t, err)

	got, err := store.GetKnowledge(ctx, "tool_result:session-1:calculator")
	require.NoError(t, err)
	require.Equal(t, entknowledge.CategoryFact, got.Category)
	require.Equal(t, "tool:calculator", got.Source)
	require.Equal(t, "{7 items}", got.Content)
}
