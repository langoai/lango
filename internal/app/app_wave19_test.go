package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agentregistry"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/receipts"
)

func TestWave19BuildMetaToolsWithRuntimesRegistersReceiptBackedToolsOnlyWithReceiptStore(t *testing.T) {
	t.Parallel()

	withoutReceipts := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, config.DefaultConfig(), nil, nil, nil, nil, nil, nil, nil)
	for _, name := range []string{
		"list_dead_lettered_post_adjudication_executions",
		"get_post_adjudication_execution_status",
		"retry_post_adjudication_execution",
	} {
		assert.Nil(t, findTool(withoutReceipts, name), "tool %q should require a receipt store", name)
	}

	withReceipts := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, config.DefaultConfig(), receipts.NewStore(), nil, nil, nil, nil, nil, nil)
	for _, name := range []string{
		"list_dead_lettered_post_adjudication_executions",
		"get_post_adjudication_execution_status",
		"retry_post_adjudication_execution",
	} {
		assert.NotNil(t, findTool(withReceipts, name), "tool %q should be registered with a receipt store", name)
	}
	assert.Nil(t, findTool(withReceipts, "execute_settlement"), "settlement execution still requires a runtime")
	assert.Nil(t, findTool(withReceipts, "execute_partial_settlement"), "partial settlement execution still requires a runtime")
}

func TestWave19ImportSkillDisabledShortCircuitsBeforeURLValidation(t *testing.T) {
	t.Parallel()

	registry := newMetaToolSkillRegistry(t)
	tool := findTool(buildMetaTools(nil, nil, registry, config.SkillConfig{AllowImport: false}, nil, nil), "import_skill")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.EqualError(t, err, "skill import disabled (skill.allowImport=false)")
}

func TestWave19InitAgentRegistryLoadsEmbeddedAndUserAgents(t *testing.T) {
	t.Parallel()

	agentsDir := t.TempDir()
	userAgentDir := filepath.Join(agentsDir, "wave19")
	require.NoError(t, os.MkdirAll(userAgentDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(userAgentDir, "AGENT.md"), []byte(`---
name: wave19
description: "Wave 19 deterministic test agent"
status: active
prefixes:
  - wave19_
---

Use deterministic test seams only.
`), 0o644))

	registry, err := initAgentRegistry(&config.AgentConfig{AgentsDir: agentsDir})
	require.NoError(t, err)

	def, ok := registry.Get("wave19")
	require.True(t, ok)
	assert.Equal(t, "Wave 19 deterministic test agent", def.Description)
	assert.Equal(t, agentregistry.SourceUser, def.Source)
	assert.Contains(t, def.Prefixes, "wave19_")

	embedded, ok := registry.Get("planner")
	require.True(t, ok, "embedded default agents should still load")
	assert.Equal(t, agentregistry.SourceEmbedded, embedded.Source)
	assert.Equal(t, agentregistry.StatusActive, embedded.Status)
}

func TestWave19NetworkModuleEnabledFollowsNetworkFeatureFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*config.Config)
		want bool
	}{
		{name: "all disabled", want: false},
		{name: "payment", edit: func(cfg *config.Config) { cfg.Payment.Enabled = true }, want: true},
		{name: "p2p", edit: func(cfg *config.Config) { cfg.P2P.Enabled = true }, want: true},
		{name: "economy", edit: func(cfg *config.Config) { cfg.Economy.Enabled = true }, want: true},
		{name: "smart account alone is not a network module trigger", edit: func(cfg *config.Config) { cfg.SmartAccount.Enabled = true }, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.DefaultConfig()
			if tt.edit != nil {
				tt.edit(cfg)
			}

			assert.Equal(t, tt.want, (&networkModule{cfg: cfg}).Enabled())
		})
	}
}

func TestWave19StopReturnsContextErrorWhenBackgroundWorkDoesNotDrain(t *testing.T) {
	app := &App{registry: lifecycle.NewRegistry()}
	app.wg.Add(1)
	defer app.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	defer cancel()

	err := app.Stop(ctx)

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled), "stop should surface wait cancellation: %v", err)
}
