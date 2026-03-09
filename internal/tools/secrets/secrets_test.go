package secrets

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/security"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSecretsTool(t *testing.T) (*Tool, *security.RefStore) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	crypto := security.NewLocalCryptoProvider()
	require.NoError(t, crypto.Initialize("test-passphrase-12345"))

	registry := security.NewKeyRegistry(client)
	ctx := context.Background()
	_, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption)
	require.NoError(t, err)

	refs := security.NewRefStore()
	store := security.NewSecretsStore(client, registry, crypto)
	return New(store, refs, nil), refs
}

func TestSecretsTool_Store(t *testing.T) {
	t.Parallel()

	tool, _ := newTestSecretsTool(t)
	ctx := context.Background()

	tests := []struct {
		give      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			give:   "store successfully",
			params: map[string]interface{}{"name": "api-key", "value": "secret-value"},
		},
		{
			give:      "empty name error",
			params:    map[string]interface{}{"name": "", "value": "secret-value"},
			wantError: true,
		},
		{
			give:      "empty value error",
			params:    map[string]interface{}{"name": "api-key", "value": ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := tool.Store(ctx, tt.params)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			m, ok := result.(map[string]interface{})
			require.True(t, ok, "expected map result, got %T", result)
			assert.Equal(t, true, m["success"])
		})
	}
}

func TestSecretsTool_Get(t *testing.T) {
	t.Parallel()

	tool, refs := newTestSecretsTool(t)
	ctx := context.Background()

	// Store a secret first
	_, err := tool.Store(ctx, map[string]interface{}{"name": "db-pass", "value": "p@ssw0rd"})
	require.NoError(t, err)

	tests := []struct {
		give      string
		params    map[string]interface{}
		wantValue string
		wantError bool
	}{
		{
			give:      "get returns reference token",
			params:    map[string]interface{}{"name": "db-pass"},
			wantValue: "{{secret:db-pass}}",
		},
		{
			give:      "non-existent secret",
			params:    map[string]interface{}{"name": "not-here"},
			wantError: true,
		},
		{
			give:      "empty name error",
			params:    map[string]interface{}{"name": ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := tool.Get(ctx, tt.params)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			m := result.(map[string]interface{})
			assert.Equal(t, tt.wantValue, m["value"])
		})
	}

	// Verify RefStore can resolve the token to actual plaintext
	t.Run("refstore resolves to plaintext", func(t *testing.T) {
		val, ok := refs.Resolve("{{secret:db-pass}}")
		require.True(t, ok, "RefStore could not resolve {{secret:db-pass}}")
		assert.Equal(t, "p@ssw0rd", string(val))
	})
}

func TestSecretsTool_List(t *testing.T) {
	t.Parallel()

	tool, _ := newTestSecretsTool(t)
	ctx := context.Background()

	t.Run("empty list count is 0", func(t *testing.T) {
		result, err := tool.List(ctx, nil)
		require.NoError(t, err)
		lr, ok := result.(ListResult)
		require.True(t, ok, "expected ListResult, got %T", result)
		assert.Equal(t, 0, lr.Count)
	})

	t.Run("store 2 then list", func(t *testing.T) {
		_, err := tool.Store(ctx, map[string]interface{}{"name": "key1", "value": "val1"})
		require.NoError(t, err)
		_, err = tool.Store(ctx, map[string]interface{}{"name": "key2", "value": "val2"})
		require.NoError(t, err)

		result, err := tool.List(ctx, nil)
		require.NoError(t, err)
		lr := result.(ListResult)
		assert.Equal(t, 2, lr.Count)
	})
}

func TestSecretsTool_Delete(t *testing.T) {
	t.Parallel()

	tool, _ := newTestSecretsTool(t)
	ctx := context.Background()

	// Store then delete
	_, err := tool.Store(ctx, map[string]interface{}{"name": "to-delete", "value": "val"})
	require.NoError(t, err)

	t.Run("delete existing", func(t *testing.T) {
		result, err := tool.Delete(ctx, map[string]interface{}{"name": "to-delete"})
		require.NoError(t, err)
		m := result.(map[string]interface{})
		assert.Equal(t, true, m["success"])
	})

	t.Run("get after delete fails", func(t *testing.T) {
		_, err := tool.Get(ctx, map[string]interface{}{"name": "to-delete"})
		require.Error(t, err)
	})

	t.Run("delete non-existent error", func(t *testing.T) {
		_, err := tool.Delete(ctx, map[string]interface{}{"name": "ghost"})
		require.Error(t, err)
	})
}

func TestSecretsTool_UpdateExisting(t *testing.T) {
	t.Parallel()

	tool, refs := newTestSecretsTool(t)
	ctx := context.Background()

	// Store initial value
	_, err := tool.Store(ctx, map[string]interface{}{"name": "mutable", "value": "v1"})
	require.NoError(t, err)

	// Store updated value with same name
	_, err = tool.Store(ctx, map[string]interface{}{"name": "mutable", "value": "v2"})
	require.NoError(t, err)

	// Get should return reference token (not plaintext)
	result, err := tool.Get(ctx, map[string]interface{}{"name": "mutable"})
	require.NoError(t, err)
	m := result.(map[string]interface{})
	assert.Equal(t, "{{secret:mutable}}", m["value"])

	// RefStore should resolve to latest value
	val, ok := refs.Resolve("{{secret:mutable}}")
	require.True(t, ok, "RefStore could not resolve {{secret:mutable}}")
	assert.Equal(t, "v2", string(val))
}
