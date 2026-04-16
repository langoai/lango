package app

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildHookRegistry_AllEnabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Hooks.Enabled = true
	cfg.Hooks.AccessControl = true
	cfg.Hooks.EventPublishing = true
	cfg.Hooks.KnowledgeSave = true

	registry := BuildHookRegistry(cfg, nil, nil)
	require.NotNil(t, registry)

	preHooks := registry.PreHooks()
	postHooks := registry.PostHooks()

	preNames := make([]string, len(preHooks))
	for i, h := range preHooks {
		preNames[i] = h.Name()
	}
	postNames := make([]string, len(postHooks))
	for i, h := range postHooks {
		postNames[i] = h.Name()
	}

	assert.Contains(t, preNames, "security_filter")
	assert.Contains(t, preNames, "agent_access_control")
	assert.Contains(t, postNames, "knowledge_save")
}

func TestBuildHookRegistry_NoBus_NoEventBusHook(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Hooks.Enabled = true
	cfg.Hooks.EventPublishing = true

	registry := BuildHookRegistry(cfg, nil, nil)

	for _, h := range registry.PreHooks() {
		assert.NotEqual(t, "eventbus", h.Name(), "EventBus hook should not be registered without a bus")
	}
}

func TestBuildHookRegistry_KnowledgeSaveDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Hooks.KnowledgeSave = false

	registry := BuildHookRegistry(cfg, nil, nil)

	for _, h := range registry.PostHooks() {
		assert.NotEqual(t, "knowledge_save", h.Name(), "KnowledgeSaveHook should not be registered when disabled")
	}
}
