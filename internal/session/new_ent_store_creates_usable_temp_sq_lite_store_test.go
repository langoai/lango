package session

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/langoai/lango/internal/ent/secret"
	entsession "github.com/langoai/lango/internal/ent/session"
	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/require"
)

func TestNewEntStore_CreatesUsableTempSQLiteStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "sessions.db")

	store, err := NewEntStore(dbPath, WithPassphrase("legacy-passphrase"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, store.Close()) })

	require.NotNil(t, store.Client())
	require.NotNil(t, store.DB())
	require.NoError(t, store.DB().Ping())

	var tableName string
	err = store.DB().QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'sessions'`,
	).Scan(&tableName)
	require.NoError(t, err)
	require.Equal(t, "sessions", tableName)

	require.NoError(t, store.Create(&Session{Key: "configCommandWiresInspectionAndMutationSubcommands3-constructor", Model: "gpt-4"}))
	got, err := store.Get("configCommandWiresInspectionAndMutationSubcommands3-constructor")
	require.NoError(t, err)
	require.Equal(t, "gpt-4", got.Model)
}

func TestEntStoreWithClient_WiresAccessorsAndOptions(t *testing.T) {
	base := newTestEntStore(t)
	wrapped := NewEntStoreWithClient(
		base.Client(),
		WithDB(base.DB()),
		WithMaxHistoryTurns(1),
		WithTTL(time.Hour),
	)

	require.Same(t, base.Client(), wrapped.Client())
	require.Same(t, base.DB(), wrapped.DB())

	require.NoError(t, wrapped.Create(&Session{Key: "configCommandWiresInspectionAndMutationSubcommands3-wrapped"}))
	require.NoError(t, wrapped.AppendMessage("configCommandWiresInspectionAndMutationSubcommands3-wrapped", Message{
		Role:      types.RoleUser,
		Content:   "older",
		Timestamp: time.Now().Add(-time.Minute),
	}))
	require.NoError(t, wrapped.AppendMessage("configCommandWiresInspectionAndMutationSubcommands3-wrapped", Message{
		Role:      types.RoleAssistant,
		Content:   "newer",
		Timestamp: time.Now(),
	}))

	got, err := wrapped.Get("configCommandWiresInspectionAndMutationSubcommands3-wrapped")
	require.NoError(t, err)
	require.Len(t, got.History, 1)
	require.Equal(t, "newer", got.History[0].Content)
}

func TestEntStore_ListSessionsOrdersByUpdatedAtDescending(t *testing.T) {
	store := newTestEntStore(t)
	ctx := context.Background()

	require.NoError(t, store.Create(&Session{Key: "configCommandWiresInspectionAndMutationSubcommands3-old"}))
	require.NoError(t, store.Create(&Session{Key: "configCommandWiresInspectionAndMutationSubcommands3-new"}))

	oldUpdated := time.Date(2026, 5, 19, 1, 0, 0, 0, time.UTC)
	newUpdated := oldUpdated.Add(time.Hour)
	setSessionUpdatedAt(t, store, "configCommandWiresInspectionAndMutationSubcommands3-old", oldUpdated)
	setSessionUpdatedAt(t, store, "configCommandWiresInspectionAndMutationSubcommands3-new", newUpdated)

	got, err := store.ListSessions(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "configCommandWiresInspectionAndMutationSubcommands3-new", got[0].Key)
	require.Equal(t, newUpdated, got[0].UpdatedAt)
	require.Equal(t, "configCommandWiresInspectionAndMutationSubcommands3-old", got[1].Key)
	require.Equal(t, oldUpdated, got[1].UpdatedAt)
	require.False(t, got[0].CreatedAt.IsZero())
}

func TestEntStore_AnnotateTimeoutPersistsSyntheticAssistantMessage(t *testing.T) {
	store := newTestEntStore(t)

	require.NoError(t, store.Create(&Session{Key: "configCommandWiresInspectionAndMutationSubcommands3-timeout"}))
	require.NoError(t, store.AnnotateTimeout("configCommandWiresInspectionAndMutationSubcommands3-timeout", "partial response"))

	got, err := store.Get("configCommandWiresInspectionAndMutationSubcommands3-timeout")
	require.NoError(t, err)
	require.Len(t, got.History, 1)
	require.Equal(t, types.RoleAssistant, got.History[0].Role)
	require.Equal(t, "[This response was interrupted due to a timeout]", got.History[0].Content)
	require.False(t, got.History[0].Timestamp.IsZero())
}

func TestEntStore_AnnotateTimeoutReturnsMissingSessionError(t *testing.T) {
	store := newTestEntStore(t)

	err := store.AnnotateTimeout("configCommandWiresInspectionAndMutationSubcommands3-missing", "partial")
	require.Error(t, err)
	require.Contains(t, err.Error(), `append message to session "configCommandWiresInspectionAndMutationSubcommands3-missing"`)
}

func TestEntStore_GetChecksumDistinguishesMissingAndUnset(t *testing.T) {
	store := newTestEntStore(t)

	_, err := store.GetChecksum("configCommandWiresInspectionAndMutationSubcommands3-missing")
	require.Error(t, err)
	require.Contains(t, err.Error(), "checksum not found: configCommandWiresInspectionAndMutationSubcommands3-missing")

	require.NoError(t, store.SetSalt("configCommandWiresInspectionAndMutationSubcommands3-unset", []byte("salt")))
	_, err = store.GetChecksum("configCommandWiresInspectionAndMutationSubcommands3-unset")
	require.Error(t, err)
	require.Contains(t, err.Error(), "checksum not set for: configCommandWiresInspectionAndMutationSubcommands3-unset")
}

func TestEntStore_GetChecksumReturnsStorageErrors(t *testing.T) {
	store := newTestEntStore(t)
	require.NoError(t, store.DB().Close())

	_, err := store.GetChecksum("configCommandWiresInspectionAndMutationSubcommands3-closed")
	require.Error(t, err)
	require.Contains(t, err.Error(), "database is closed")
}

func TestEntStore_MigrateSecretsReencryptsAndUpdatesSecurityConfig(t *testing.T) {
	store := newTestEntStore(t)
	ctx := context.Background()

	createNewEntStoreCreatesUsableTempSqLiteStoreSecrets(t, store, "configCommandWiresInspectionAndMutationSubcommands3-success", map[string][]byte{
		"configCommandWiresInspectionAndMutationSubcommands3-first":  []byte("first-old"),
		"configCommandWiresInspectionAndMutationSubcommands3-second": []byte("second-old"),
	})
	require.NoError(t, store.SetSalt("default", []byte("old-salt")))
	require.NoError(t, store.SetChecksum("default", []byte("old-checksum")))

	var migrated [][]byte
	err := store.MigrateSecrets(ctx, func(ciphertext []byte) ([]byte, error) {
		migrated = append(migrated, append([]byte(nil), ciphertext...))
		return append([]byte("new:"), ciphertext...), nil
	}, []byte("new-salt"), []byte("new-checksum"))
	require.NoError(t, err)
	require.ElementsMatch(t, [][]byte{[]byte("first-old"), []byte("second-old")}, migrated)

	first := getSecretValue(t, store, "configCommandWiresInspectionAndMutationSubcommands3-first")
	second := getSecretValue(t, store, "configCommandWiresInspectionAndMutationSubcommands3-second")
	require.Equal(t, []byte("new:first-old"), first)
	require.Equal(t, []byte("new:second-old"), second)

	salt, err := store.GetSalt("default")
	require.NoError(t, err)
	require.Equal(t, []byte("new-salt"), salt)
	checksum, err := store.GetChecksum("default")
	require.NoError(t, err)
	require.Equal(t, []byte("new-checksum"), checksum)
}

func TestEntStore_MigrateSecretsReencryptErrorRollsBack(t *testing.T) {
	store := newTestEntStore(t)
	ctx := context.Background()

	createNewEntStoreCreatesUsableTempSqLiteStoreSecrets(t, store, "configCommandWiresInspectionAndMutationSubcommands3-error", map[string][]byte{
		"configCommandWiresInspectionAndMutationSubcommands3-error-first": []byte("first-old"),
	})
	require.NoError(t, store.SetSalt("default", []byte("old-salt")))
	require.NoError(t, store.SetChecksum("default", []byte("old-checksum")))

	errReencrypt := errors.New("reencrypt failed")
	err := store.MigrateSecrets(ctx, func([]byte) ([]byte, error) {
		return nil, errReencrypt
	}, []byte("new-salt"), []byte("new-checksum"))
	require.Error(t, err)
	require.ErrorIs(t, err, errReencrypt)
	require.Contains(t, err.Error(), "re-encrypt secret configCommandWiresInspectionAndMutationSubcommands3-error-first")

	require.Equal(t, []byte("first-old"), getSecretValue(t, store, "configCommandWiresInspectionAndMutationSubcommands3-error-first"))
	salt, err := store.GetSalt("default")
	require.NoError(t, err)
	require.Equal(t, []byte("old-salt"), salt)
	checksum, err := store.GetChecksum("default")
	require.NoError(t, err)
	require.Equal(t, []byte("old-checksum"), checksum)
}

func TestEntStore_MigrateSecretsReportsMissingSecurityConfigTable(t *testing.T) {
	store := newTestEntStore(t)
	ctx := context.Background()

	createNewEntStoreCreatesUsableTempSqLiteStoreSecrets(t, store, "configCommandWiresInspectionAndMutationSubcommands3-missing-config", map[string][]byte{
		"configCommandWiresInspectionAndMutationSubcommands3-missing-config-secret": []byte("old"),
	})
	require.NoError(t, dropSecurityConfigTable(store.DB()))

	err := store.MigrateSecrets(ctx, func(ciphertext []byte) ([]byte, error) {
		return append([]byte("new:"), ciphertext...), nil
	}, []byte("new-salt"), []byte("new-checksum"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "update salt")

	require.Equal(t, []byte("old"), getSecretValue(t, store, "configCommandWiresInspectionAndMutationSubcommands3-missing-config-secret"))
}

func setSessionUpdatedAt(t *testing.T, store *EntStore, key string, updatedAt time.Time) {
	t.Helper()

	ctx := context.Background()
	entSession, err := store.client.Session.Query().Where(entsession.Key(key)).Only(ctx)
	require.NoError(t, err)
	_, err = store.client.Session.UpdateOne(entSession).SetUpdatedAt(updatedAt).Save(ctx)
	require.NoError(t, err)
}

func createNewEntStoreCreatesUsableTempSqLiteStoreSecrets(
	t *testing.T,
	store *EntStore,
	keyName string,
	values map[string][]byte,
) {
	t.Helper()

	ctx := context.Background()
	key, err := store.client.Key.Create().
		SetName(keyName).
		SetRemoteKeyID("local").
		SetType("encryption").
		Save(ctx)
	require.NoError(t, err)

	for name, value := range values {
		_, err := store.client.Secret.Create().
			SetName(name).
			SetEncryptedValue(value).
			SetKey(key).
			Save(ctx)
		require.NoError(t, err)
	}
}

func getSecretValue(t *testing.T, store *EntStore, name string) []byte {
	t.Helper()

	got, err := store.client.Secret.Query().Where(secret.Name(name)).Only(context.Background())
	require.NoError(t, err)
	return got.EncryptedValue
}

func dropSecurityConfigTable(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS security_config (
		name TEXT PRIMARY KEY,
		value BLOB NOT NULL,
		checksum BLOB
	)`); err != nil {
		return err
	}
	_, err := db.Exec(`DROP TABLE security_config`)
	return err
}
