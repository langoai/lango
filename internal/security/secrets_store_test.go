package security

import (
	"context"
	"errors"
	"testing"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// testCryptoProvider is a simple in-memory encrypt/decrypt mock for testing
// SecretsStore without depending on real crypto.
type testCryptoProvider struct {
	prefix     string
	encryptErr error
	decryptErr error
}

func (p *testCryptoProvider) Sign(_ context.Context, _ string, payload []byte) ([]byte, error) {
	return append([]byte("sig:"), payload...), nil
}

func (p *testCryptoProvider) Encrypt(_ context.Context, _ string, plaintext []byte) ([]byte, error) {
	if p.encryptErr != nil {
		return nil, p.encryptErr
	}
	// Prefix plaintext so we can verify round-trip
	return append([]byte(p.prefix), plaintext...), nil
}

func (p *testCryptoProvider) Decrypt(_ context.Context, _ string, ciphertext []byte) ([]byte, error) {
	if p.decryptErr != nil {
		return nil, p.decryptErr
	}
	prefix := []byte(p.prefix)
	if len(ciphertext) < len(prefix) {
		return nil, errors.New("invalid ciphertext")
	}
	return ciphertext[len(prefix):], nil
}

// newTestSecretsStore sets up a SecretsStore with an in-memory DB,
// a KeyRegistry pre-seeded with a default encryption key, and a mock CryptoProvider.
func newTestSecretsStore(t *testing.T) (*SecretsStore, *KeyRegistry) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	registry := NewKeyRegistry(client)
	ctx := context.Background()

	// Register a default encryption key
	_, err := registry.RegisterKey(ctx, "default-enc", "local", KeyTypeEncryption)
	require.NoError(t, err)

	crypto := &testCryptoProvider{prefix: "ENC:"}
	store := NewSecretsStore(client, registry, crypto)
	return store, registry
}

func TestSecretsStore_Store(t *testing.T) {
	t.Parallel()

	store, _ := newTestSecretsStore(t)
	ctx := context.Background()

	tests := []struct {
		give    string
		name    string
		value   []byte
		wantErr bool
	}{
		{
			give:  "store a new secret",
			name:  "api-key",
			value: []byte("sk-12345"),
		},
		{
			give:  "store another secret",
			name:  "db-password",
			value: []byte("p@ssw0rd"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := store.Store(ctx, tt.name, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestSecretsStore_Store_UpdateExisting(t *testing.T) {
	t.Parallel()

	store, _ := newTestSecretsStore(t)
	ctx := context.Background()

	// Store initial value
	err := store.Store(ctx, "mutable-secret", []byte("v1"))
	require.NoError(t, err)

	// Store again with same name (should update, not duplicate)
	err = store.Store(ctx, "mutable-secret", []byte("v2"))
	require.NoError(t, err)

	// Retrieve and verify updated value
	val, err := store.Get(ctx, "mutable-secret")
	require.NoError(t, err)
	assert.Equal(t, []byte("v2"), val)

	// List should show exactly 1 secret
	secrets, err := store.List(ctx)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
}

func TestSecretsStore_Get(t *testing.T) {
	t.Parallel()

	store, _ := newTestSecretsStore(t)
	ctx := context.Background()

	// Seed a secret
	err := store.Store(ctx, "get-me", []byte("secret-data"))
	require.NoError(t, err)

	tests := []struct {
		give    string
		name    string
		want    []byte
		wantErr bool
	}{
		{
			give: "existing secret",
			name: "get-me",
			want: []byte("secret-data"),
		},
		{
			give:    "non-existent secret",
			name:    "ghost",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			val, err := store.Get(ctx, tt.name)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, val)
		})
	}
}

func TestSecretsStore_Get_IncrementsAccessCount(t *testing.T) {
	t.Parallel()

	store, _ := newTestSecretsStore(t)
	ctx := context.Background()

	err := store.Store(ctx, "counted", []byte("val"))
	require.NoError(t, err)

	// Access twice
	_, err = store.Get(ctx, "counted")
	require.NoError(t, err)
	_, err = store.Get(ctx, "counted")
	require.NoError(t, err)

	// Check access count via List
	secrets, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	assert.Equal(t, 2, secrets[0].AccessCount)
}

func TestSecretsStore_Get_UpdatesKeyLastUsed(t *testing.T) {
	t.Parallel()

	store, registry := newTestSecretsStore(t)
	ctx := context.Background()

	err := store.Store(ctx, "lu-test", []byte("val"))
	require.NoError(t, err)

	// Access secret
	_, err = store.Get(ctx, "lu-test")
	require.NoError(t, err)

	// Verify the key's last_used_at is now set
	keyInfo, err := registry.GetKey(ctx, "default-enc")
	require.NoError(t, err)
	assert.NotNil(t, keyInfo.LastUsedAt, "key last_used_at should be set after secret access")
}

func TestSecretsStore_List(t *testing.T) {
	t.Parallel()

	store, _ := newTestSecretsStore(t)
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		secrets, err := store.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, secrets)
	})

	t.Run("returns all secrets with metadata", func(t *testing.T) {
		err := store.Store(ctx, "secret-a", []byte("a"))
		require.NoError(t, err)
		err = store.Store(ctx, "secret-b", []byte("b"))
		require.NoError(t, err)

		secrets, err := store.List(ctx)
		require.NoError(t, err)
		assert.Len(t, secrets, 2)

		// Verify metadata fields are populated
		for _, s := range secrets {
			assert.NotZero(t, s.ID)
			assert.NotEmpty(t, s.Name)
			assert.NotZero(t, s.CreatedAt)
			assert.NotZero(t, s.UpdatedAt)
			assert.NotZero(t, s.KeyID)
			assert.Equal(t, "default-enc", s.KeyName)
			assert.Equal(t, 0, s.AccessCount)
		}
	})
}

func TestSecretsStore_Delete(t *testing.T) {
	t.Parallel()

	store, _ := newTestSecretsStore(t)
	ctx := context.Background()

	// Seed a secret
	err := store.Store(ctx, "to-delete", []byte("val"))
	require.NoError(t, err)

	tests := []struct {
		give    string
		name    string
		wantErr bool
	}{
		{
			give: "delete existing secret",
			name: "to-delete",
		},
		{
			give:    "delete non-existent secret",
			name:    "no-such-secret",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := store.Delete(ctx, tt.name)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestSecretsStore_Delete_ThenGetFails(t *testing.T) {
	t.Parallel()

	store, _ := newTestSecretsStore(t)
	ctx := context.Background()

	err := store.Store(ctx, "ephemeral", []byte("val"))
	require.NoError(t, err)

	err = store.Delete(ctx, "ephemeral")
	require.NoError(t, err)

	_, err = store.Get(ctx, "ephemeral")
	require.Error(t, err)
}

func TestSecretsStore_Store_NoEncryptionKey(t *testing.T) {
	t.Parallel()

	// Create store with no encryption key registered
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	registry := NewKeyRegistry(client)
	crypto := &testCryptoProvider{prefix: "ENC:"}
	store := NewSecretsStore(client, registry, crypto)

	err := store.Store(context.Background(), "orphan", []byte("val"))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoEncryptionKeys)
}

func TestSecretsStore_Store_EncryptError(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	registry := NewKeyRegistry(client)
	ctx := context.Background()
	_, err := registry.RegisterKey(ctx, "enc-key", "local", KeyTypeEncryption)
	require.NoError(t, err)

	crypto := &testCryptoProvider{
		prefix:     "ENC:",
		encryptErr: errors.New("hw failure"),
	}
	store := NewSecretsStore(client, registry, crypto)

	err = store.Store(ctx, "broken", []byte("val"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encrypt secret")
}

func TestSecretsStore_Get_DecryptError(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	registry := NewKeyRegistry(client)
	ctx := context.Background()
	_, err := registry.RegisterKey(ctx, "enc-key", "local", KeyTypeEncryption)
	require.NoError(t, err)

	// Use a working crypto for Store, then swap to failing one for Get
	goodCrypto := &testCryptoProvider{prefix: "ENC:"}
	store := NewSecretsStore(client, registry, goodCrypto)

	err = store.Store(ctx, "will-fail", []byte("val"))
	require.NoError(t, err)

	// Replace crypto with one that fails decryption
	store.crypto = &testCryptoProvider{
		prefix:     "ENC:",
		decryptErr: errors.New("tampered"),
	}

	_, err = store.Get(ctx, "will-fail")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt secret")
}

func TestSecretsStore_FullCRUDCycle(t *testing.T) {
	t.Parallel()

	store, _ := newTestSecretsStore(t)
	ctx := context.Background()

	// Create
	err := store.Store(ctx, "lifecycle", []byte("initial"))
	require.NoError(t, err)

	// Read
	val, err := store.Get(ctx, "lifecycle")
	require.NoError(t, err)
	assert.Equal(t, []byte("initial"), val)

	// Update
	err = store.Store(ctx, "lifecycle", []byte("updated"))
	require.NoError(t, err)

	val, err = store.Get(ctx, "lifecycle")
	require.NoError(t, err)
	assert.Equal(t, []byte("updated"), val)

	// List should have exactly 1
	secrets, err := store.List(ctx)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "lifecycle", secrets[0].Name)

	// Delete
	err = store.Delete(ctx, "lifecycle")
	require.NoError(t, err)

	// Verify gone
	_, err = store.Get(ctx, "lifecycle")
	require.Error(t, err)

	// List should be empty
	secrets, err = store.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, secrets)
}
