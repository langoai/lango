package app

import (
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	require.NotNil(t, app.Agent, "expected agent to be initialized")
	require.NotNil(t, app.Gateway, "expected gateway to be initialized")
	require.NotNil(t, app.Store, "expected store to be initialized")
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
	require.NoError(t, err, "New() should succeed without security config")
}

func TestNew_NoProviders(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	_, err := New(testBoot(t, cfg))
	require.Error(t, err, "expected error when no providers configured")
}

func TestNew_InvalidProviderType(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"test": {Type: "nonexistent", APIKey: "test-key"},
	}
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	_, err := New(testBoot(t, cfg))
	require.Error(t, err, "expected error for invalid provider type")
}
