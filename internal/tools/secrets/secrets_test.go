package secrets

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/security"
	_ "github.com/mattn/go-sqlite3"
)

func newTestSecretsTool(t *testing.T) (*Tool, *security.RefStore) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	crypto := security.NewLocalCryptoProvider()
	if err := crypto.Initialize("test-passphrase-12345"); err != nil {
		t.Fatalf("initialize crypto: %v", err)
	}

	registry := security.NewKeyRegistry(client)
	ctx := context.Background()
	if _, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption); err != nil {
		t.Fatalf("register key: %v", err)
	}

	refs := security.NewRefStore()
	store := security.NewSecretsStore(client, registry, crypto)
	return New(store, refs, nil), refs
}

func TestSecretsTool_Store(t *testing.T) {
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
			if m["success"] != true {
				t.Error("expected success=true")
			}
		})
	}
}

func TestSecretsTool_Get(t *testing.T) {
	tool, refs := newTestSecretsTool(t)
	ctx := context.Background()

	// Store a secret first
	_, err := tool.Store(ctx, map[string]interface{}{"name": "db-pass", "value": "p@ssw0rd"})
	if err != nil {
		t.Fatalf("store: %v", err)
	}

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
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			m := result.(map[string]interface{})
			if m["value"] != tt.wantValue {
				t.Errorf("value: want %q, got %q", tt.wantValue, m["value"])
			}
		})
	}

	// Verify RefStore can resolve the token to actual plaintext
	t.Run("refstore resolves to plaintext", func(t *testing.T) {
		val, ok := refs.Resolve("{{secret:db-pass}}")
		if !ok {
			t.Fatal("RefStore could not resolve {{secret:db-pass}}")
		}
		if string(val) != "p@ssw0rd" {
			t.Errorf("resolved value: want %q, got %q", "p@ssw0rd", val)
		}
	})
}

func TestSecretsTool_List(t *testing.T) {
	tool, _ := newTestSecretsTool(t)
	ctx := context.Background()

	t.Run("empty list count is 0", func(t *testing.T) {
		result, err := tool.List(ctx, nil)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		lr, ok := result.(ListResult)
		if !ok {
			t.Fatalf("expected ListResult, got %T", result)
		}
		if lr.Count != 0 {
			t.Errorf("count: want 0, got %d", lr.Count)
		}
	})

	t.Run("store 2 then list", func(t *testing.T) {
		if _, err := tool.Store(ctx, map[string]interface{}{"name": "key1", "value": "val1"}); err != nil {
			t.Fatalf("store key1: %v", err)
		}
		if _, err := tool.Store(ctx, map[string]interface{}{"name": "key2", "value": "val2"}); err != nil {
			t.Fatalf("store key2: %v", err)
		}

		result, err := tool.List(ctx, nil)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		lr := result.(ListResult)
		if lr.Count != 2 {
			t.Errorf("count: want 2, got %d", lr.Count)
		}
	})
}

func TestSecretsTool_Delete(t *testing.T) {
	tool, _ := newTestSecretsTool(t)
	ctx := context.Background()

	// Store then delete
	if _, err := tool.Store(ctx, map[string]interface{}{"name": "to-delete", "value": "val"}); err != nil {
		t.Fatalf("store: %v", err)
	}

	t.Run("delete existing", func(t *testing.T) {
		result, err := tool.Delete(ctx, map[string]interface{}{"name": "to-delete"})
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		m := result.(map[string]interface{})
		if m["success"] != true {
			t.Error("expected success=true")
		}
	})

	t.Run("get after delete fails", func(t *testing.T) {
		_, err := tool.Get(ctx, map[string]interface{}{"name": "to-delete"})
		if err == nil {
			t.Fatal("expected error for deleted secret")
		}
	})

	t.Run("delete non-existent error", func(t *testing.T) {
		_, err := tool.Delete(ctx, map[string]interface{}{"name": "ghost"})
		if err == nil {
			t.Fatal("expected error for non-existent secret")
		}
	})
}

func TestSecretsTool_UpdateExisting(t *testing.T) {
	tool, refs := newTestSecretsTool(t)
	ctx := context.Background()

	// Store initial value
	if _, err := tool.Store(ctx, map[string]interface{}{"name": "mutable", "value": "v1"}); err != nil {
		t.Fatalf("store v1: %v", err)
	}

	// Store updated value with same name
	if _, err := tool.Store(ctx, map[string]interface{}{"name": "mutable", "value": "v2"}); err != nil {
		t.Fatalf("store v2: %v", err)
	}

	// Get should return reference token (not plaintext)
	result, err := tool.Get(ctx, map[string]interface{}{"name": "mutable"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	m := result.(map[string]interface{})
	if m["value"] != "{{secret:mutable}}" {
		t.Errorf("value: want %q, got %q", "{{secret:mutable}}", m["value"])
	}

	// RefStore should resolve to latest value
	val, ok := refs.Resolve("{{secret:mutable}}")
	if !ok {
		t.Fatal("RefStore could not resolve {{secret:mutable}}")
	}
	if string(val) != "v2" {
		t.Errorf("resolved value: want %q, got %q", "v2", val)
	}
}
