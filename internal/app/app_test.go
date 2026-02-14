package app

import (
	"path/filepath"
	"testing"

	"github.com/langowarny/lango/internal/bootstrap"
	"github.com/langowarny/lango/internal/config"
)

// testBoot creates a minimal bootstrap.Result for testing.
func testBoot(t *testing.T, cfg *config.Config) *bootstrap.Result {
	t.Helper()
	return &bootstrap.Result{
		Config: cfg,
	}
}

func TestNew_MinimalConfig(t *testing.T) {
	t.Skip("requires provider credentials; run manually with GOOGLE_API_KEY set")

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "google"
	cfg.Agent.Model = "gemini-2.0-flash"
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	app, err := New(testBoot(t, cfg))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if app.Agent == nil {
		t.Fatal("expected agent to be initialized")
	}
	if app.Gateway == nil {
		t.Fatal("expected gateway to be initialized")
	}
	if app.Store == nil {
		t.Fatal("expected store to be initialized")
	}
}

func TestNew_SecurityDisabledByDefault(t *testing.T) {
	t.Skip("requires provider credentials; run manually with GOOGLE_API_KEY set")

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "google"
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	// Security is not configured — should not block startup
	_, err := New(testBoot(t, cfg))
	if err != nil {
		t.Fatalf("New() should succeed without security config, got: %v", err)
	}
}

func TestNew_NoProviders(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	_, err := New(testBoot(t, cfg))
	if err == nil {
		t.Fatal("expected error when no providers configured")
	}
}

func TestNew_InvalidProviderType(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"test": {Type: "nonexistent", APIKey: "test-key"},
	}
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	_, err := New(testBoot(t, cfg))
	if err == nil {
		t.Fatal("expected error for invalid provider type")
	}
}
