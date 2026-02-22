package configstore

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/security"
)

func testClient(t *testing.T) *ent.Client {
	t.Helper()
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	err = client.Schema.Create(context.Background())
	require.NoError(t, err)
	return client
}

func testCrypto(t *testing.T, passphrase string) *security.LocalCryptoProvider {
	t.Helper()
	crypto := security.NewLocalCryptoProvider()
	err := crypto.Initialize(passphrase)
	require.NoError(t, err)
	return crypto
}

func testConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Host:             "localhost",
			Port:             8080,
			HTTPEnabled:      true,
			WebSocketEnabled: false,
		},
		Agent: config.AgentConfig{
			Provider:    "anthropic",
			Model:       "claude-sonnet-4-20250514",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "console",
		},
		Session: config.SessionConfig{
			DatabasePath:    "~/.lango/sessions.db",
			TTL:             24 * time.Hour,
			MaxHistoryTurns: 50,
		},
	}
}

func TestStore_SaveLoadRoundTrip(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	original := testConfig()

	err := store.Save(ctx, "default", original)
	require.NoError(t, err)

	loaded, err := store.Load(ctx, "default")
	require.NoError(t, err)

	assert.Equal(t, original.Server.Host, loaded.Server.Host)
	assert.Equal(t, original.Server.Port, loaded.Server.Port)
	assert.Equal(t, original.Agent.Provider, loaded.Agent.Provider)
	assert.Equal(t, original.Agent.Model, loaded.Agent.Model)
	assert.Equal(t, original.Agent.MaxTokens, loaded.Agent.MaxTokens)
	assert.Equal(t, original.Logging.Level, loaded.Logging.Level)
}

func TestStore_SaveUpdatesExisting(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	cfg := testConfig()

	err := store.Save(ctx, "default", cfg)
	require.NoError(t, err)

	cfg.Server.Port = 9090
	err = store.Save(ctx, "default", cfg)
	require.NoError(t, err)

	loaded, err := store.Load(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, 9090, loaded.Server.Port)

	// Version should have incremented.
	profiles, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Equal(t, 2, profiles[0].Version)
}

func TestStore_LoadNotFound(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	_, err := store.Load(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrProfileNotFound)
}

func TestStore_List(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	cfg := testConfig()

	err := store.Save(ctx, "alpha", cfg)
	require.NoError(t, err)
	err = store.Save(ctx, "beta", cfg)
	require.NoError(t, err)
	err = store.Save(ctx, "gamma", cfg)
	require.NoError(t, err)

	profiles, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, profiles, 3)

	// Should be ordered by name.
	assert.Equal(t, "alpha", profiles[0].Name)
	assert.Equal(t, "beta", profiles[1].Name)
	assert.Equal(t, "gamma", profiles[2].Name)
}

func TestStore_Delete(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	cfg := testConfig()

	err := store.Save(ctx, "to-delete", cfg)
	require.NoError(t, err)

	exists, err := store.Exists(ctx, "to-delete")
	require.NoError(t, err)
	assert.True(t, exists)

	err = store.Delete(ctx, "to-delete")
	require.NoError(t, err)

	exists, err = store.Exists(ctx, "to-delete")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestStore_DeleteNotFound(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	err := store.Delete(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrProfileNotFound)
}

func TestStore_DeleteActiveProfile(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	cfg := testConfig()
	err := store.Save(ctx, "active-profile", cfg)
	require.NoError(t, err)
	err = store.SetActive(ctx, "active-profile")
	require.NoError(t, err)

	err = store.Delete(ctx, "active-profile")
	assert.ErrorIs(t, err, ErrDeleteActive)
}

func TestStore_SetActive(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	cfg := testConfig()

	err := store.Save(ctx, "first", cfg)
	require.NoError(t, err)
	err = store.Save(ctx, "second", cfg)
	require.NoError(t, err)

	// Activate first.
	err = store.SetActive(ctx, "first")
	require.NoError(t, err)

	profiles, err := store.List(ctx)
	require.NoError(t, err)
	for _, p := range profiles {
		if p.Name == "first" {
			assert.True(t, p.Active)
		} else {
			assert.False(t, p.Active)
		}
	}

	// Switch to second.
	err = store.SetActive(ctx, "second")
	require.NoError(t, err)

	profiles, err = store.List(ctx)
	require.NoError(t, err)
	for _, p := range profiles {
		if p.Name == "second" {
			assert.True(t, p.Active)
		} else {
			assert.False(t, p.Active)
		}
	}
}

func TestStore_SetActiveNotFound(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	err := store.SetActive(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrProfileNotFound)
}

func TestStore_LoadActive(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	cfg := testConfig()
	err := store.Save(ctx, "my-profile", cfg)
	require.NoError(t, err)
	err = store.SetActive(ctx, "my-profile")
	require.NoError(t, err)

	name, loaded, err := store.LoadActive(ctx)
	require.NoError(t, err)
	assert.Equal(t, "my-profile", name)
	assert.Equal(t, cfg.Server.Port, loaded.Server.Port)
}

func TestStore_LoadActiveNoActive(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	_, _, err := store.LoadActive(ctx)
	assert.ErrorIs(t, err, ErrNoActiveProfile)
}

func TestStore_WrongPassphrase(t *testing.T) {
	client := testClient(t)
	crypto1 := testCrypto(t, "correct-passphrase")
	store1 := NewStore(client, crypto1)
	ctx := context.Background()

	cfg := testConfig()
	err := store1.Save(ctx, "encrypted", cfg)
	require.NoError(t, err)

	// Try to load with a different passphrase.
	crypto2 := testCrypto(t, "wrong-passphrase-here")
	store2 := NewStore(client, crypto2)

	_, err = store2.Load(ctx, "encrypted")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt")
}

func TestStore_Exists(t *testing.T) {
	client := testClient(t)
	crypto := testCrypto(t, "test-passphrase-123")
	store := NewStore(client, crypto)
	ctx := context.Background()

	exists, err := store.Exists(ctx, "nope")
	require.NoError(t, err)
	assert.False(t, exists)

	cfg := testConfig()
	err = store.Save(ctx, "yes", cfg)
	require.NoError(t, err)

	exists, err = store.Exists(ctx, "yes")
	require.NoError(t, err)
	assert.True(t, exists)
}
