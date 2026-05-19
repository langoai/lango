package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
)

func TestWave46InitAgentRegistryIgnoresInvalidUserAgentsDir(t *testing.T) {
	t.Parallel()

	agentsDir := filepath.Join(t.TempDir(), "agents-file")
	require.NoError(t, os.WriteFile(agentsDir, []byte("not a directory"), 0o600))

	registry, err := initAgentRegistry(&config.AgentConfig{AgentsDir: agentsDir})

	require.NoError(t, err)
	require.NotNil(t, registry)
	assert.NotEmpty(t, registry.All(), "embedded agents should remain available")
}

func TestWave46NetworkModuleInitAllNetworkFeaturesDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = false
	cfg.Economy.Enabled = false
	cfg.SmartAccount.Enabled = false

	result, err := (&networkModule{cfg: cfg}).Init(
		context.Background(),
		staticResolver{appinit.ProvidesSupervisor: &foundationValues{}},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.Nil(t, result.Values[appinit.ProvidesEconomy])
	assert.Nil(t, result.Values[appinit.ProvidesContract])
	assert.Nil(t, result.Values[appinit.ProvidesSmartAccount])
	assert.Nil(t, result.Values[appinit.ProvidesWorkspace])

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "contract").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.Equal(t, "P2P networking (disabled)", p2pEntry.Description)
	assert.Nil(t, wave46CatalogEntry(result.CatalogEntries, "workspace"))
}

func TestWave46CreateSkillRejectsMalformedJSONBeforeDependencies(t *testing.T) {
	t.Parallel()

	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil), "create_skill")
	require.NotNil(t, tool)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "definition",
			params: map[string]interface{}{
				"name":        "wave46-json",
				"description": "JSON validation",
				"type":        "template",
				"definition":  `{not-json`,
			},
			wantError: "parse definition JSON",
		},
		{
			name: "parameters",
			params: map[string]interface{}{
				"name":        "wave46-params",
				"description": "Parameter validation",
				"type":        "template",
				"definition":  `{}`,
				"parameters":  `{not-json`,
			},
			wantError: "parse parameters JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tool.Handler(context.Background(), tt.params)

			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestWave46InitP2PRejectsInvalidListenAddressWithoutRuntimeStartup(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "keys")
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/not-a-port"}

	assert.Nil(t, initP2P(cfg, &wiringP2PWallet{}, nil, nil, nil, nil, nil, nil, ""))
}

func TestWave46PathInsideDirRejectsMixedAbsoluteRelativeInputs(t *testing.T) {
	t.Parallel()

	absoluteRoot := t.TempDir()

	assert.False(t, pathInsideDir("relative/file.txt", absoluteRoot))
	assert.False(t, pathInsideDir(filepath.Join(absoluteRoot, "file.txt"), "relative-root"))
}

func wave46CatalogEntry(entries []appinit.CatalogEntry, category string) *appinit.CatalogEntry {
	for i := range entries {
		if entries[i].Category == category {
			return &entries[i]
		}
	}
	return nil
}
