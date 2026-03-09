package crypto

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/security"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCryptoProvider struct {
	encryptFn func(ctx context.Context, keyID string, plaintext []byte) ([]byte, error)
	decryptFn func(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error)
	signFn    func(ctx context.Context, keyID string, payload []byte) ([]byte, error)
}

func (m *mockCryptoProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	return m.encryptFn(ctx, keyID, plaintext)
}

func (m *mockCryptoProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	return m.decryptFn(ctx, keyID, ciphertext)
}

func (m *mockCryptoProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	return m.signFn(ctx, keyID, payload)
}

func TestCryptoTool_Hash(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	mock := &mockCryptoProvider{}
	registry := security.NewKeyRegistry(client)
	refs := security.NewRefStore()
	tool := New(mock, registry, refs, nil)
	ctx := context.Background()

	// Compute expected sha256 hash of "hello"
	sum := sha256.Sum256([]byte("hello"))
	wantSHA256 := hex.EncodeToString(sum[:])

	tests := []struct {
		give      string
		params    map[string]interface{}
		wantHash  string
		wantAlgo  string
		wantError bool
	}{
		{
			give:     "sha256 known value",
			params:   map[string]interface{}{"data": "hello", "algorithm": "sha256"},
			wantHash: wantSHA256,
			wantAlgo: "sha256",
		},
		{
			give:     "sha512",
			params:   map[string]interface{}{"data": "hello", "algorithm": "sha512"},
			wantAlgo: "sha512",
		},
		{
			give:     "default algorithm is sha256",
			params:   map[string]interface{}{"data": "hello"},
			wantHash: wantSHA256,
			wantAlgo: "sha256",
		},
		{
			give:      "unsupported algorithm md5",
			params:    map[string]interface{}{"data": "hello", "algorithm": "md5"},
			wantError: true,
		},
		{
			give:      "empty data error",
			params:    map[string]interface{}{"data": ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := tool.Hash(ctx, tt.params)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			m, ok := result.(map[string]interface{})
			require.True(t, ok, "expected map result, got %T", result)
			assert.Equal(t, tt.wantAlgo, m["algorithm"])
			if tt.wantHash != "" {
				assert.Equal(t, tt.wantHash, m["hash"])
			}
		})
	}
}

func TestCryptoTool_Encrypt(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	registry := security.NewKeyRegistry(client)
	_, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption)
	require.NoError(t, err)

	refs := security.NewRefStore()

	// Mock returns reversed bytes
	mock := &mockCryptoProvider{
		encryptFn: func(_ context.Context, _ string, plaintext []byte) ([]byte, error) {
			reversed := make([]byte, len(plaintext))
			for i, b := range plaintext {
				reversed[len(plaintext)-1-i] = b
			}
			return reversed, nil
		},
	}
	tool := New(mock, registry, refs, nil)

	tests := []struct {
		give      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			give:   "encrypt success",
			params: map[string]interface{}{"data": "hello"},
		},
		{
			give:      "empty data error",
			params:    map[string]interface{}{"data": ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := tool.Encrypt(ctx, tt.params)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			m, ok := result.(map[string]interface{})
			require.True(t, ok, "expected map result, got %T", result)
			ciphertext, ok := m["ciphertext"].(string)
			require.True(t, ok, "expected ciphertext to be string")
			// Verify it's valid base64
			decoded, err := base64.StdEncoding.DecodeString(ciphertext)
			require.NoError(t, err, "ciphertext is not valid base64")
			// Mock reverses bytes, so decoded should be reversed "hello"
			assert.Equal(t, "olleh", string(decoded))
		})
	}

	// Provider returns error
	t.Run("provider error", func(t *testing.T) {
		errMock := &mockCryptoProvider{
			encryptFn: func(_ context.Context, _ string, _ []byte) ([]byte, error) {
				return nil, fmt.Errorf("provider failure")
			},
		}
		errTool := New(errMock, registry, refs, nil)
		_, err := errTool.Encrypt(ctx, map[string]interface{}{"data": "hello"})
		require.Error(t, err)
	})
}

func TestCryptoTool_Decrypt(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	registry := security.NewKeyRegistry(client)
	_, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption)
	require.NoError(t, err)

	refs := security.NewRefStore()

	// Mock: encrypt reverses, decrypt reverses back
	mock := &mockCryptoProvider{
		encryptFn: func(_ context.Context, _ string, plaintext []byte) ([]byte, error) {
			reversed := make([]byte, len(plaintext))
			for i, b := range plaintext {
				reversed[len(plaintext)-1-i] = b
			}
			return reversed, nil
		},
		decryptFn: func(_ context.Context, _ string, ciphertext []byte) ([]byte, error) {
			reversed := make([]byte, len(ciphertext))
			for i, b := range ciphertext {
				reversed[len(ciphertext)-1-i] = b
			}
			return reversed, nil
		},
	}
	tool := New(mock, registry, refs, nil)

	t.Run("decrypt returns reference token", func(t *testing.T) {
		encResult, err := tool.Encrypt(ctx, map[string]interface{}{"data": "secret"})
		require.NoError(t, err)
		encMap := encResult.(map[string]interface{})
		ciphertext := encMap["ciphertext"].(string)

		decResult, err := tool.Decrypt(ctx, map[string]interface{}{"ciphertext": ciphertext})
		require.NoError(t, err)
		decMap := decResult.(map[string]interface{})

		// Value should be a reference token, not plaintext
		dataStr, ok := decMap["data"].(string)
		require.True(t, ok, "expected data to be string, got %T", decMap["data"])
		assert.True(t, strings.HasPrefix(dataStr, "{{decrypt:"))
		assert.True(t, strings.HasSuffix(dataStr, "}}"))

		// RefStore should resolve the token to actual plaintext
		val, ok := refs.Resolve(dataStr)
		require.True(t, ok, "RefStore could not resolve %q", dataStr)
		assert.Equal(t, "secret", string(val))
	})

	t.Run("empty ciphertext error", func(t *testing.T) {
		_, err := tool.Decrypt(ctx, map[string]interface{}{"ciphertext": ""})
		require.Error(t, err)
	})

	t.Run("invalid base64 error", func(t *testing.T) {
		_, err := tool.Decrypt(ctx, map[string]interface{}{"ciphertext": "not-valid-base64!!!"})
		require.Error(t, err)
	})

	t.Run("provider error", func(t *testing.T) {
		errMock := &mockCryptoProvider{
			decryptFn: func(_ context.Context, _ string, _ []byte) ([]byte, error) {
				return nil, fmt.Errorf("decrypt failure")
			},
		}
		errTool := New(errMock, registry, refs, nil)
		validB64 := base64.StdEncoding.EncodeToString([]byte("data"))
		_, err := errTool.Decrypt(ctx, map[string]interface{}{"ciphertext": validB64})
		require.Error(t, err)
	})
}

func TestCryptoTool_Sign(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	registry := security.NewKeyRegistry(client)
	refs := security.NewRefStore()

	fixedSig := []byte("fixed-signature-bytes")
	mock := &mockCryptoProvider{
		signFn: func(_ context.Context, _ string, _ []byte) ([]byte, error) {
			return fixedSig, nil
		},
	}
	tool := New(mock, registry, refs, nil)

	tests := []struct {
		give      string
		params    map[string]interface{}
		wantSig   string
		wantError bool
	}{
		{
			give:    "sign with explicit keyId",
			params:  map[string]interface{}{"data": "hello", "keyId": "my-key"},
			wantSig: base64.StdEncoding.EncodeToString(fixedSig),
		},
		{
			give:    "default keyId is local",
			params:  map[string]interface{}{"data": "hello"},
			wantSig: base64.StdEncoding.EncodeToString(fixedSig),
		},
		{
			give:      "empty data error",
			params:    map[string]interface{}{"data": ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := tool.Sign(ctx, tt.params)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			m := result.(map[string]interface{})
			assert.Equal(t, tt.wantSig, m["signature"])
		})
	}
}

func TestCryptoTool_Keys(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	registry := security.NewKeyRegistry(client)
	refs := security.NewRefStore()
	mock := &mockCryptoProvider{}
	tool := New(mock, registry, refs, nil)

	// Register 2 keys
	_, err := registry.RegisterKey(ctx, "key1", "remote1", security.KeyTypeEncryption)
	require.NoError(t, err)
	_, err = registry.RegisterKey(ctx, "key2", "remote2", security.KeyTypeSigning)
	require.NoError(t, err)

	result, err := tool.Keys(ctx, nil)
	require.NoError(t, err)
	m := result.(map[string]interface{})
	count, ok := m["count"].(int)
	require.True(t, ok, "expected count to be int, got %T", m["count"])
	assert.Equal(t, 2, count)
}

func TestMapToStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		input     map[string]interface{}
		wantData  string
		wantAlgo  string
		wantError bool
	}{
		{
			give:     "valid map to HashParams",
			input:    map[string]interface{}{"data": "hello", "algorithm": "sha256"},
			wantData: "hello",
			wantAlgo: "sha256",
		},
		{
			give:      "type mismatch (number for string field) returns error",
			input:     map[string]interface{}{"data": 123},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			var p HashParams
			err := mapToStruct(tt.input, &p)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantData != "" {
				assert.Equal(t, tt.wantData, p.Data)
			}
			if tt.wantAlgo != "" {
				assert.Equal(t, tt.wantAlgo, p.Algorithm)
			}
		})
	}
}
