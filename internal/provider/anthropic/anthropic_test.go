package anthropic

import (
	"context"
	"testing"
)

func TestNewProvider(t *testing.T) {
	p := NewProvider("my-anthropic", "test-key")
	if p.ID() != "my-anthropic" {
		t.Errorf("expected ID 'my-anthropic', got %s", p.ID())
	}
}

func TestAnthropicProvider_ListModels(t *testing.T) {
	p := NewProvider("anthropic", "test-key")
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected at least one model")
	}
	// Verify known models exist
	found := false
	for _, m := range models {
		if m.ID == "claude-3-5-sonnet-latest" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected claude-3-5-sonnet-latest in model list")
	}
}
