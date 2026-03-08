package security

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func newTestKeyRegistry(t *testing.T) *KeyRegistry {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	return NewKeyRegistry(client)
}

func TestKeyType_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give KeyType
		want bool
	}{
		{give: KeyTypeEncryption, want: true},
		{give: KeyTypeSigning, want: true},
		{give: KeyType("unknown"), want: false},
		{give: KeyType(""), want: false},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.Valid())
		})
	}
}

func TestKeyType_Values(t *testing.T) {
	t.Parallel()

	vals := KeyType("").Values()
	assert.Contains(t, vals, KeyTypeEncryption)
	assert.Contains(t, vals, KeyTypeSigning)
	assert.Len(t, vals, 2)
}

func TestKeyRegistry_RegisterKey(t *testing.T) {
	t.Parallel()

	reg := newTestKeyRegistry(t)
	ctx := context.Background()

	tests := []struct {
		give        string
		name        string
		remoteKeyID string
		keyType     KeyType
		wantErr     bool
	}{
		{
			give:        "register encryption key",
			name:        "enc-key-1",
			remoteKeyID: "remote-enc-1",
			keyType:     KeyTypeEncryption,
		},
		{
			give:        "register signing key",
			name:        "sign-key-1",
			remoteKeyID: "remote-sign-1",
			keyType:     KeyTypeSigning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			info, err := reg.RegisterKey(ctx, tt.name, tt.remoteKeyID, tt.keyType)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.name, info.Name)
			assert.Equal(t, tt.remoteKeyID, info.RemoteKeyID)
			assert.Equal(t, tt.keyType, info.Type)
			assert.NotZero(t, info.ID)
			assert.NotZero(t, info.CreatedAt)
			assert.Nil(t, info.LastUsedAt)
		})
	}
}

func TestKeyRegistry_RegisterKey_UpdateExisting(t *testing.T) {
	t.Parallel()

	reg := newTestKeyRegistry(t)
	ctx := context.Background()

	// Register initial key
	info1, err := reg.RegisterKey(ctx, "my-key", "remote-1", KeyTypeEncryption)
	require.NoError(t, err)

	// Re-register with same name updates the key
	info2, err := reg.RegisterKey(ctx, "my-key", "remote-2", KeyTypeSigning)
	require.NoError(t, err)

	assert.Equal(t, info1.ID, info2.ID, "ID should remain the same on update")
	assert.Equal(t, "remote-2", info2.RemoteKeyID)
	assert.Equal(t, KeyTypeSigning, info2.Type)
}

func TestKeyRegistry_GetKey(t *testing.T) {
	t.Parallel()

	reg := newTestKeyRegistry(t)
	ctx := context.Background()

	// Seed a key
	_, err := reg.RegisterKey(ctx, "get-test", "remote-get", KeyTypeEncryption)
	require.NoError(t, err)

	tests := []struct {
		give    string
		name    string
		wantErr bool
	}{
		{
			give: "existing key",
			name: "get-test",
		},
		{
			give:    "non-existent key",
			name:    "no-such-key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			info, err := reg.GetKey(ctx, tt.name)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrKeyNotFound)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.name, info.Name)
			assert.Equal(t, "remote-get", info.RemoteKeyID)
			assert.Equal(t, KeyTypeEncryption, info.Type)
		})
	}
}

func TestKeyRegistry_GetDefaultKey(t *testing.T) {
	t.Parallel()

	reg := newTestKeyRegistry(t)
	ctx := context.Background()

	t.Run("no encryption keys returns error", func(t *testing.T) {
		_, err := reg.GetDefaultKey(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoEncryptionKeys)
	})

	t.Run("returns most recent encryption key", func(t *testing.T) {
		_, err := reg.RegisterKey(ctx, "enc-old", "remote-old", KeyTypeEncryption)
		require.NoError(t, err)

		_, err = reg.RegisterKey(ctx, "enc-new", "remote-new", KeyTypeEncryption)
		require.NoError(t, err)

		// Register a signing key to ensure it is not returned
		_, err = reg.RegisterKey(ctx, "sign-only", "remote-sign", KeyTypeSigning)
		require.NoError(t, err)

		info, err := reg.GetDefaultKey(ctx)
		require.NoError(t, err)
		assert.Equal(t, "enc-new", info.Name)
		assert.Equal(t, KeyTypeEncryption, info.Type)
	})
}

func TestKeyRegistry_ListKeys(t *testing.T) {
	t.Parallel()

	reg := newTestKeyRegistry(t)
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		keys, err := reg.ListKeys(ctx)
		require.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("returns all keys ordered by created_at desc", func(t *testing.T) {
		_, err := reg.RegisterKey(ctx, "key-a", "r-a", KeyTypeEncryption)
		require.NoError(t, err)

		_, err = reg.RegisterKey(ctx, "key-b", "r-b", KeyTypeSigning)
		require.NoError(t, err)

		_, err = reg.RegisterKey(ctx, "key-c", "r-c", KeyTypeEncryption)
		require.NoError(t, err)

		keys, err := reg.ListKeys(ctx)
		require.NoError(t, err)
		assert.Len(t, keys, 3)

		// Most recently created should be first
		assert.Equal(t, "key-c", keys[0].Name)
	})
}

func TestKeyRegistry_UpdateLastUsed(t *testing.T) {
	t.Parallel()

	reg := newTestKeyRegistry(t)
	ctx := context.Background()

	_, err := reg.RegisterKey(ctx, "update-lu", "remote-lu", KeyTypeEncryption)
	require.NoError(t, err)

	// Initially last_used_at is nil
	info, err := reg.GetKey(ctx, "update-lu")
	require.NoError(t, err)
	assert.Nil(t, info.LastUsedAt)

	// Update last used
	err = reg.UpdateLastUsed(ctx, "update-lu")
	require.NoError(t, err)

	// Verify last_used_at is now set
	info, err = reg.GetKey(ctx, "update-lu")
	require.NoError(t, err)
	assert.NotNil(t, info.LastUsedAt)
}

func TestKeyRegistry_DeleteKey(t *testing.T) {
	t.Parallel()

	reg := newTestKeyRegistry(t)
	ctx := context.Background()

	_, err := reg.RegisterKey(ctx, "to-delete", "remote-del", KeyTypeEncryption)
	require.NoError(t, err)

	t.Run("delete existing key", func(t *testing.T) {
		err := reg.DeleteKey(ctx, "to-delete")
		require.NoError(t, err)

		// Verify it is gone
		_, err = reg.GetKey(ctx, "to-delete")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("delete non-existent key does not error", func(t *testing.T) {
		// ent Delete returns 0 affected rows but no error
		err := reg.DeleteKey(ctx, "no-such-key")
		require.NoError(t, err)
	})
}

func TestKeyRegistry_FullCRUDCycle(t *testing.T) {
	t.Parallel()

	reg := newTestKeyRegistry(t)
	ctx := context.Background()

	// Create
	info, err := reg.RegisterKey(ctx, "lifecycle", "remote-lc", KeyTypeEncryption)
	require.NoError(t, err)
	assert.Equal(t, "lifecycle", info.Name)

	// Read
	got, err := reg.GetKey(ctx, "lifecycle")
	require.NoError(t, err)
	assert.Equal(t, info.ID, got.ID)

	// Update (re-register with same name)
	updated, err := reg.RegisterKey(ctx, "lifecycle", "remote-lc-v2", KeyTypeSigning)
	require.NoError(t, err)
	assert.Equal(t, info.ID, updated.ID)
	assert.Equal(t, "remote-lc-v2", updated.RemoteKeyID)
	assert.Equal(t, KeyTypeSigning, updated.Type)

	// List should show exactly 1 key
	keys, err := reg.ListKeys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 1)

	// Delete
	err = reg.DeleteKey(ctx, "lifecycle")
	require.NoError(t, err)

	// Verify gone
	_, err = reg.GetKey(ctx, "lifecycle")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrKeyNotFound)

	// List should be empty
	keys, err = reg.ListKeys(ctx)
	require.NoError(t, err)
	assert.Empty(t, keys)
}
