package agent

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRegistryOutput_PreAndPost(t *testing.T) {
	t.Parallel()

	registry := toolchain.NewHookRegistry()
	registry.RegisterPre(toolchain.NewSecurityFilterHook(nil))
	registry.RegisterPost(toolchain.NewKnowledgeSaveHook(nil, toolchain.DefaultSaveableTools))

	hCfg := config.HooksConfig{}
	out := buildRegistryOutput(registry, hCfg)

	require.Len(t, out.PreHooks, 1)
	assert.Equal(t, "security_filter", out.PreHooks[0].Name)
	assert.Equal(t, "pre", out.PreHooks[0].Phase)
	assert.True(t, out.PreHooks[0].Wirable)

	require.Len(t, out.PostHooks, 1)
	assert.Equal(t, "knowledge_save", out.PostHooks[0].Name)
	assert.Equal(t, "post", out.PostHooks[0].Phase)

	details := out.PostHooks[0].Details
	require.NotNil(t, details)
	tools, ok := details["saveableTools"]
	require.True(t, ok)
	toolList, ok := tools.([]string)
	require.True(t, ok)
	assert.NotEmpty(t, toolList)
}

func TestBuildRegistryOutput_Empty(t *testing.T) {
	t.Parallel()

	registry := toolchain.NewHookRegistry()
	hCfg := config.HooksConfig{}
	out := buildRegistryOutput(registry, hCfg)

	assert.Empty(t, out.PreHooks)
	assert.Empty(t, out.PostHooks)
}

func TestBuildRegistryOutput_EventBusPlaceholder(t *testing.T) {
	t.Parallel()

	registry := toolchain.NewHookRegistry()
	registry.RegisterPre(toolchain.NewSecurityFilterHook(nil))

	hCfg := config.HooksConfig{EventPublishing: true}
	out := buildRegistryOutput(registry, hCfg)

	var found bool
	for _, hi := range out.PreHooks {
		if hi.Name == "eventbus" {
			found = true
			assert.False(t, hi.Wirable)
			assert.Contains(t, hi.Reason, "event bus")
			assert.Equal(t, "pre+post", hi.Phase)
		}
	}
	assert.True(t, found, "EventBus placeholder should be present when EventPublishing is enabled")
}

func TestPrintJSON_BackwardCompatible(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Hooks.Enabled = true
	cfg.Hooks.SecurityFilter = true
	cfg.Hooks.KnowledgeSave = true

	h := cfg.Hooks
	registry := toolchain.NewHookRegistry()
	registry.RegisterPre(toolchain.NewSecurityFilterHook(nil))

	out := fullOutput{
		hooksConfigOutput: hooksConfigOutput{
			Enabled:         h.Enabled,
			SecurityFilter:  h.SecurityFilter,
			AccessControl:   h.AccessControl,
			EventPublishing: h.EventPublishing,
			KnowledgeSave:   h.KnowledgeSave,
			BlockedCommands: h.BlockedCommands,
		},
		Registry: buildRegistryOutput(registry, h),
	}

	assert.True(t, out.Enabled)
	assert.True(t, out.SecurityFilter)
	assert.NotNil(t, out.Registry)
	assert.Len(t, out.Registry.PreHooks, 1)
}
