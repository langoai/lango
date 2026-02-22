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
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			m, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("expected map result, got %T", result)
			}
			if m["algorithm"] != tt.wantAlgo {
				t.Errorf("algorithm: want %s, got %s", tt.wantAlgo, m["algorithm"])
			}
			if tt.wantHash != "" && m["hash"] != tt.wantHash {
				t.Errorf("hash: want %s, got %s", tt.wantHash, m["hash"])
			}
		})
	}
}

func TestCryptoTool_Encrypt(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	registry := security.NewKeyRegistry(client)
	if _, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption); err != nil {
		t.Fatalf("register key: %v", err)
	}

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
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			m, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("expected map result, got %T", result)
			}
			ciphertext, ok := m["ciphertext"].(string)
			if !ok {
				t.Fatal("expected ciphertext to be string")
			}
			// Verify it's valid base64
			decoded, err := base64.StdEncoding.DecodeString(ciphertext)
			if err != nil {
				t.Fatalf("ciphertext is not valid base64: %v", err)
			}
			// Mock reverses bytes, so decoded should be reversed "hello"
			want := "olleh"
			if string(decoded) != want {
				t.Errorf("decoded ciphertext: want %q, got %q", want, string(decoded))
			}
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
		if err == nil {
			t.Fatal("expected error from provider failure")
		}
	})
}

func TestCryptoTool_Decrypt(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	registry := security.NewKeyRegistry(client)
	if _, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption); err != nil {
		t.Fatalf("register key: %v", err)
	}

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
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		encMap := encResult.(map[string]interface{})
		ciphertext := encMap["ciphertext"].(string)

		decResult, err := tool.Decrypt(ctx, map[string]interface{}{"ciphertext": ciphertext})
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}
		decMap := decResult.(map[string]interface{})

		// Value should be a reference token, not plaintext
		dataStr, ok := decMap["data"].(string)
		if !ok {
			t.Fatalf("expected data to be string, got %T", decMap["data"])
		}
		if !strings.HasPrefix(dataStr, "{{decrypt:") || !strings.HasSuffix(dataStr, "}}") {
			t.Errorf("expected reference token {{decrypt:...}}, got %q", dataStr)
		}

		// RefStore should resolve the token to actual plaintext
		val, ok := refs.Resolve(dataStr)
		if !ok {
			t.Fatalf("RefStore could not resolve %q", dataStr)
		}
		if string(val) != "secret" {
			t.Errorf("resolved value: want %q, got %q", "secret", val)
		}
	})

	t.Run("empty ciphertext error", func(t *testing.T) {
		_, err := tool.Decrypt(ctx, map[string]interface{}{"ciphertext": ""})
		if err == nil {
			t.Fatal("expected error for empty ciphertext")
		}
	})

	t.Run("invalid base64 error", func(t *testing.T) {
		_, err := tool.Decrypt(ctx, map[string]interface{}{"ciphertext": "not-valid-base64!!!"})
		if err == nil {
			t.Fatal("expected error for invalid base64")
		}
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
		if err == nil {
			t.Fatal("expected error from provider failure")
		}
	})
}

func TestCryptoTool_Sign(t *testing.T) {
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
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			m := result.(map[string]interface{})
			if m["signature"] != tt.wantSig {
				t.Errorf("signature: want %s, got %s", tt.wantSig, m["signature"])
			}
		})
	}
}

func TestCryptoTool_Keys(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	registry := security.NewKeyRegistry(client)
	refs := security.NewRefStore()
	mock := &mockCryptoProvider{}
	tool := New(mock, registry, refs, nil)

	// Register 2 keys
	if _, err := registry.RegisterKey(ctx, "key1", "remote1", security.KeyTypeEncryption); err != nil {
		t.Fatalf("register key1: %v", err)
	}
	if _, err := registry.RegisterKey(ctx, "key2", "remote2", security.KeyTypeSigning); err != nil {
		t.Fatalf("register key2: %v", err)
	}

	result, err := tool.Keys(ctx, nil)
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	m := result.(map[string]interface{})
	count, ok := m["count"].(int)
	if !ok {
		t.Fatalf("expected count to be int, got %T", m["count"])
	}
	if count != 2 {
		t.Errorf("count: want 2, got %d", count)
	}
}

func TestMapToStruct(t *testing.T) {
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
			var p HashParams
			err := mapToStruct(tt.input, &p)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantData != "" && p.Data != tt.wantData {
				t.Errorf("Data: want %q, got %q", tt.wantData, p.Data)
			}
			if tt.wantAlgo != "" && p.Algorithm != tt.wantAlgo {
				t.Errorf("Algorithm: want %q, got %q", tt.wantAlgo, p.Algorithm)
			}
		})
	}
}
