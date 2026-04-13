package app

import (
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
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

// ── Phase B Cleanup Stack Tests ──

func TestCleanupStack_RollbackReverseOrder(t *testing.T) {
	var order []string
	var s cleanupStack

	s.push("step-A", func() { order = append(order, "A") })
	s.push("step-B", func() { order = append(order, "B") })
	s.push("step-C", func() { order = append(order, "C") })

	s.rollback()

	assert.Equal(t, []string{"C", "B", "A"}, order, "cleanups must execute in reverse order")
	assert.Empty(t, s.entries, "stack should be empty after rollback")
}

func TestCleanupStack_ClearDiscardsWithoutExecution(t *testing.T) {
	executed := false
	var s cleanupStack

	s.push("step-A", func() { executed = true })
	s.clear()

	assert.False(t, executed, "clear must not execute cleanup functions")
	assert.Empty(t, s.entries, "stack should be empty after clear")
}

func TestCleanupStack_RollbackEmpty(t *testing.T) {
	var s cleanupStack
	// Should not panic on empty stack.
	s.rollback()
	assert.Empty(t, s.entries)
}

func TestCleanupStack_PushAndRollbackPartial(t *testing.T) {
	var order []string
	var s cleanupStack

	s.push("output-store", func() { order = append(order, "output-store") })
	s.push("gateway", func() { order = append(order, "gateway") })

	// Simulate B6 failure — rollback should clean up gateway then output-store.
	s.rollback()

	assert.Equal(t, []string{"gateway", "output-store"}, order,
		"B6 failure should rollback gateway then output-store")
}

func TestNew_PhaseBRollback_AgentCreationFailure(t *testing.T) {
	// No providers configured: Phase A succeeds (supervisor with zero providers),
	// but B6 (initAgent) fails — triggering Phase B rollback of OutputStore + Gateway.
	cfg := config.DefaultConfig()
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")

	_, err := New(testBoot(t, cfg))
	require.Error(t, err, "expected error when agent creation fails")
}
