package agent

import (
	"encoding/json"
	"io"
	"os"
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

func TestAgentHooksJSONSnapshotShape(t *testing.T) {
	// Not parallel: captureStdout swaps os.Stdout, which is process-global.
	cfg := config.DefaultConfig()
	cfg.Hooks.Enabled = true
	cfg.Hooks.SecurityFilter = true
	cfg.Hooks.AccessControl = true
	cfg.Hooks.EventPublishing = true
	cfg.Hooks.KnowledgeSave = true
	cfg.Hooks.BlockedCommands = []string{"rm -rf /"}

	cmd := newHooksCmd(func() (*config.Config, error) {
		return cfg, nil
	})
	cmd.SetArgs([]string{"--json"})

	output, err := captureStdout(t, cmd.Execute)
	require.NoError(t, err)

	var out fullOutput
	require.NoError(t, json.Unmarshal([]byte(output), &out))

	assert.True(t, out.Enabled)
	assert.True(t, out.SecurityFilter)
	assert.True(t, out.AccessControl)
	assert.True(t, out.EventPublishing)
	assert.True(t, out.KnowledgeSave)
	assert.Equal(t, []string{"rm -rf /"}, out.BlockedCommands)

	securityFilter, ok := findHookInfo(out.Registry.PreHooks, "security_filter")
	require.True(t, ok)
	assert.Equal(t, "pre", securityFilter.Phase)
	assert.True(t, securityFilter.Wirable)

	accessControl, ok := findHookInfo(out.Registry.PreHooks, "agent_access_control")
	require.True(t, ok)
	assert.Equal(t, "pre", accessControl.Phase)
	assert.True(t, accessControl.Wirable)

	eventBus, ok := findHookInfo(out.Registry.PreHooks, "eventbus")
	require.True(t, ok)
	assert.Equal(t, "pre+post", eventBus.Phase)
	assert.False(t, eventBus.Wirable)
	assert.Contains(t, eventBus.Reason, "event bus")

	knowledgeSave, ok := findHookInfo(out.Registry.PostHooks, "knowledge_save")
	require.True(t, ok)
	assert.Equal(t, "post", knowledgeSave.Phase)
	assert.True(t, knowledgeSave.Wirable)
	require.NotNil(t, knowledgeSave.Details)
	assert.Equal(t, "fallback-constant", knowledgeSave.Details["source"])
	saveableTools, ok := knowledgeSave.Details["saveableTools"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, saveableTools)
}

func captureStdout(t *testing.T, run func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	defer reader.Close()

	writerClosed := false
	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
		if !writerClosed {
			_ = writer.Close()
		}
	}()

	runErr := run()
	closeErr := writer.Close()
	writerClosed = true

	data, readErr := io.ReadAll(reader)
	require.NoError(t, readErr)
	if runErr != nil {
		return string(data), runErr
	}
	return string(data), closeErr
}

func findHookInfo(hooks []hookInfo, name string) (hookInfo, bool) {
	for _, hook := range hooks {
		if hook.Name == name {
			return hook, true
		}
	}
	return hookInfo{}, false
}
