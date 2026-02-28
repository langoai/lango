package anthropic

import (
	"context"
	"os"
	"testing"
)

func TestNewProvider(t *testing.T) {
	p := NewProvider("my-anthropic", "test-key")
	if p.ID() != "my-anthropic" {
		t.Errorf("expected ID 'my-anthropic', got %s", p.ID())
	}
}

func TestAnthropicProvider_ListModels(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set; skipping live API test")
	}

	p := NewProvider("anthropic", apiKey)
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected at least one model")
	}
	// Verify the API returns model IDs
	for _, m := range models {
		if m.ID == "" {
			t.Error("expected non-empty model ID")
		}
	}
}
