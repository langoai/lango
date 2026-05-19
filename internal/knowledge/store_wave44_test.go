package knowledge

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/security"
)

type wave44FailingPayloadProtector struct{}

func (wave44FailingPayloadProtector) EncryptPayload([]byte) ([]byte, []byte, int, error) {
	return nil, nil, 0, errors.New("wave44 encrypt failed")
}

func (wave44FailingPayloadProtector) DecryptPayload([]byte, []byte, int) ([]byte, error) {
	return nil, errors.New("wave44 decrypt failed")
}

func TestWave44PayloadProtectorErrorsSurfaceFromKnowledgeAndLearningWrites(t *testing.T) {
	store := newTestStore(t)
	store.SetPayloadProtector(wave44FailingPayloadProtector{})
	ctx := context.Background()

	err := store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave44-knowledge-encrypt-error",
		Category: entknowledge.CategoryFact,
		Content:  "payload should fail before persistence",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "prepare knowledge payload")
	require.Contains(t, err.Error(), "wave44 encrypt failed")

	err = store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave44 learning encrypt error",
		ErrorPattern: "payload should fail before learning persistence",
		Diagnosis:    "encryption failure",
		Fix:          "surface error",
		Category:     entlearning.CategoryToolError,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "prepare learning payload")
	require.Contains(t, err.Error(), "wave44 encrypt failed")
}

func TestWave44KnowledgeDecryptFailureFallsBackToStoredProjection(t *testing.T) {
	store := newTestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	const key = "wave44-knowledge-decrypt-fallback"
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      key,
		Category: entknowledge.CategoryFact,
		Content:  "email alice@example.com token SECRETSECRETSECRETSECRETSECRETSECRET",
	}))

	row := store.client.Knowledge.Query().Where(entknowledge.Key(key)).OnlyX(ctx)
	require.NotNil(t, row.ContentKeyVersion)
	_, err := row.Update().SetContentKeyVersion(security.PayloadKeyVersionV1 + 99).Save(ctx)
	require.NoError(t, err)

	got, err := store.GetKnowledge(ctx, key)
	require.NoError(t, err)
	require.NotContains(t, got.Content, "alice@example.com")
	require.NotContains(t, got.Content, "SECRETSECRETSECRETSECRETSECRETSECRET")
	require.Contains(t, got.Content, "[email]")
	require.Contains(t, got.Content, "[secret]")
}

func TestWave44LearningDecryptFailureIsReturnedByReadsAndSearch(t *testing.T) {
	store := newTestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave44 corrupt learning",
		ErrorPattern: "context deadline exceeded",
		Diagnosis:    "timeout diagnosis",
		Fix:          "increase timeout",
		Category:     entlearning.CategoryTimeout,
	}))
	row := store.client.Learning.Query().OnlyX(ctx)
	require.NotNil(t, row.PayloadKeyVersion)
	_, err := row.Update().SetPayloadKeyVersion(security.PayloadKeyVersionV1 + 99).Save(ctx)
	require.NoError(t, err)

	_, err = store.GetLearning(ctx, row.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "decrypt learning payload")

	_, err = store.SearchLearnings(ctx, "deadline", "", 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "resolve learning")
	require.Contains(t, err.Error(), "decrypt learning payload")
}

func TestWave44FTS5SyncFailuresDoNotBlockEntKnowledgeWrites(t *testing.T) {
	store, _ := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.fts5Index.DropTable())
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave44-fts-sync-failure",
		Category: entknowledge.CategoryFact,
		Content:  "knowledge survives fts sync failure",
	}))

	got, err := store.GetKnowledge(ctx, "wave44-fts-sync-failure")
	require.NoError(t, err)
	require.Equal(t, "knowledge survives fts sync failure", got.Content)

	results, err := store.SearchKnowledge(ctx, "survives", "", 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "wave44-fts-sync-failure", results[0].Key)

	scored, err := store.SearchKnowledgeScored(ctx, "survives", "", 10)
	require.NoError(t, err)
	require.Len(t, scored, 1)
	require.Equal(t, "wave44-fts-sync-failure", scored[0].Entry.Key)
	require.Equal(t, "like", scored[0].SearchSource)
}

func TestWave44LearningFTS5SearchFailureFallsBackToLIKE(t *testing.T) {
	store, _ := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave44 learning fallback",
		ErrorPattern: "fallback learning search",
		Diagnosis:    "fts table dropped after save",
		Fix:          "use like fallback",
		Category:     entlearning.CategoryToolError,
	}))
	require.NoError(t, store.learningFTS5Idx.DropTable())

	got, err := store.SearchLearnings(ctx, "fallback", "", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "wave44 learning fallback", got[0].Trigger)
}
