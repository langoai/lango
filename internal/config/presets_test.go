package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidPreset(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{"minimal", true},
		{"researcher", true},
		{"collaborator", true},
		{"full", true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidPreset(tt.give))
		})
	}
}

func TestPresetConfig_Minimal(t *testing.T) {
	cfg := PresetConfig("minimal")
	require.NotNil(t, cfg)

	// Minimal should match DefaultConfig.
	def := DefaultConfig()
	assert.Equal(t, def.Knowledge.Enabled, cfg.Knowledge.Enabled)
	assert.Equal(t, def.Graph.Enabled, cfg.Graph.Enabled)
}

func TestPresetConfig_Researcher(t *testing.T) {
	cfg := PresetConfig("researcher")
	require.NotNil(t, cfg)

	assert.True(t, cfg.Knowledge.Enabled)
	assert.True(t, cfg.ObservationalMemory.Enabled)
	assert.True(t, cfg.Graph.Enabled)
	assert.Equal(t, "openai", cfg.Embedding.Provider)
	assert.True(t, cfg.Librarian.Enabled)
	// Should not enable unrelated features.
	assert.False(t, cfg.P2P.Enabled)
	assert.False(t, cfg.Payment.Enabled)
}

func TestPresetConfig_Collaborator(t *testing.T) {
	cfg := PresetConfig("collaborator")
	require.NotNil(t, cfg)

	assert.True(t, cfg.P2P.Enabled)
	assert.True(t, cfg.Payment.Enabled)
	assert.Equal(t, "https://sepolia.base.org", cfg.Payment.Network.RPCURL)
	assert.True(t, cfg.Economy.Enabled)
	// Should not enable knowledge features.
	assert.False(t, cfg.Knowledge.Enabled)
}

func TestPresetConfig_Full(t *testing.T) {
	cfg := PresetConfig("full")
	require.NotNil(t, cfg)

	assert.True(t, cfg.Knowledge.Enabled)
	assert.True(t, cfg.ObservationalMemory.Enabled)
	assert.True(t, cfg.Graph.Enabled)
	assert.True(t, cfg.Librarian.Enabled)
	assert.True(t, cfg.Cron.Enabled)
	assert.True(t, cfg.Background.Enabled)
	assert.True(t, cfg.Workflow.Enabled)
	assert.True(t, cfg.MCP.Enabled)
	assert.True(t, cfg.AgentMemory.Enabled)
	assert.True(t, cfg.Agent.MultiAgent)
}

func TestPresetConfig_Unknown(t *testing.T) {
	cfg := PresetConfig("nonexistent")
	require.NotNil(t, cfg)

	def := DefaultConfig()
	assert.Equal(t, def.Knowledge.Enabled, cfg.Knowledge.Enabled)
}

func TestAllPresets(t *testing.T) {
	presets := AllPresets()
	assert.Len(t, presets, 4)

	for _, p := range presets {
		assert.NotEmpty(t, p.Name)
		assert.NotEmpty(t, p.Desc)
		assert.True(t, IsValidPreset(string(p.Name)))
	}
}
