package app

import (
	"testing"

	"github.com/langoai/lango/internal/config"
)

func TestInitSupervisor(t *testing.T) {
	t.Skip("requires provider credentials")

	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	sv, err := initSupervisor(cfg)
	if err != nil {
		t.Fatalf("initSupervisor() returned error: %v", err)
	}
	if sv == nil {
		t.Fatal("expected supervisor to be initialized")
	}
}
