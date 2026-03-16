package app

import (
	"testing"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModuleTopoSort_AllDisabled verifies that when all optional modules are disabled,
// the build succeeds with only the foundation module.
func TestModuleTopoSort_AllDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	modules := []appinit.Module{
		&foundationModule{cfg: cfg},
		&intelligenceModule{cfg: cfg},
		&automationModule{cfg: cfg},
		&networkModule{cfg: cfg},
		&extensionModule{cfg: cfg},
	}

	sorted, err := appinit.TopoSort(modules)
	require.NoError(t, err)
	require.NotEmpty(t, sorted)

	// Foundation should come first (no dependencies).
	assert.Equal(t, "foundation", sorted[0].Name())
}

// TestModuleTopoSort_DependencyOrder verifies that the intelligence module
// comes after the foundation module.
func TestModuleTopoSort_DependencyOrder(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Knowledge.Enabled = true

	modules := []appinit.Module{
		&intelligenceModule{cfg: cfg},
		&foundationModule{cfg: cfg},
		&automationModule{cfg: cfg},
		&extensionModule{cfg: cfg},
	}

	sorted, err := appinit.TopoSort(modules)
	require.NoError(t, err)

	names := make([]string, len(sorted))
	for i, m := range sorted {
		names[i] = m.Name()
	}

	// Foundation must come before intelligence.
	foundIdx := indexOf(names, "foundation")
	intelIdx := indexOf(names, "intelligence")
	assert.True(t, foundIdx < intelIdx, "foundation should come before intelligence: %v", names)
}

// TestModuleEnabled_Automation verifies that the automation module is disabled
// when all automation subsystems are disabled.
func TestModuleEnabled_Automation(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	// All disabled by default.
	m := &automationModule{cfg: cfg}
	assert.False(t, m.Enabled())

	cfg2 := config.DefaultConfig()
	cfg2.Cron.Enabled = true
	m2 := &automationModule{cfg: cfg2}
	assert.True(t, m2.Enabled())
}

// TestModuleBuild_FoundationOnly verifies that foundation module initializes
// successfully when other modules are disabled.
func TestModuleBuild_FoundationOnly(t *testing.T) {
	// This test would require a bootstrap.Result which needs DB setup.
	// Skipping for unit tests — validated in integration tests.
	t.Skip("requires bootstrap.Result with DB client")
}

// TestFoundationCatalogEntries verifies catalog entry generation.
func TestFoundationCatalogEntries(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	entries := buildFoundationCatalogEntries(cfg, nil, nil, nil)

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Category] = true
	}

	assert.True(t, names["exec"])
	assert.True(t, names["filesystem"])
	assert.True(t, names["browser"])
	assert.True(t, names["crypto"])
	assert.True(t, names["secrets"])
}

func indexOf(s []string, target string) int {
	for i, v := range s {
		if v == target {
			return i
		}
	}
	return -1
}

// TestModuleBuild_DisabledModuleDependency verifies that disabled modules
// don't block the initialization of modules that depend on them.
func TestModuleBuild_DisabledModuleDependency(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	// Network depends on Security/SessionStore.
	// When network is disabled, the build should still succeed.
	modules := []appinit.Module{
		&foundationModule{cfg: cfg},
		&networkModule{cfg: cfg}, // disabled (payment/p2p/economy all false)
	}

	sorted, err := appinit.TopoSort(modules)
	require.NoError(t, err)
	// Only foundation should be in sorted (network is disabled).
	require.Len(t, sorted, 1)
	assert.Equal(t, "foundation", sorted[0].Name())
}

// TestExtensionModule_AlwaysEnabled verifies that the extension module is
// always enabled.
func TestExtensionModule_AlwaysEnabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	m := &extensionModule{cfg: cfg}
	assert.True(t, m.Enabled())
}

// TestIntelligenceModule_AlwaysEnabled verifies that the intelligence module is
// always enabled (individual subsystems check their own config).
func TestIntelligenceModule_AlwaysEnabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	m := &intelligenceModule{cfg: cfg}
	assert.True(t, m.Enabled())
}

// TestModuleProvides verifies that each module declares its provides keys correctly.
func TestModuleProvides(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	tests := []struct {
		name     string
		module   appinit.Module
		wantKeys []appinit.Provides
	}{
		{
			name:     "foundation",
			module:   &foundationModule{cfg: cfg},
			wantKeys: []appinit.Provides{appinit.ProvidesSupervisor, appinit.ProvidesSessionStore, appinit.ProvidesSecurity},
		},
		{
			name:   "intelligence",
			module: &intelligenceModule{cfg: cfg},
			wantKeys: []appinit.Provides{
				appinit.ProvidesKnowledge, appinit.ProvidesMemory,
				appinit.ProvidesEmbedding, appinit.ProvidesGraph,
				appinit.ProvidesLibrarian, appinit.ProvidesSkills,
			},
		},
		{
			name:     "automation",
			module:   &automationModule{cfg: cfg},
			wantKeys: []appinit.Provides{appinit.ProvidesAutomation},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantKeys, tt.module.Provides())
		})
	}
}
