package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/types"
)

func TestResolveContextAutoEnable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		giveCfg      func() *Config
		giveExplicit map[string]bool
		wantKnow     bool
		wantMem      bool
		wantRetr     bool
		wantAutoKnow bool
		wantAutoMem  bool
		wantAutoRetr bool
	}{
		{
			give: "nil explicitKeys + dbPath → auto-enable all",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Session.DatabasePath = "/tmp/test.db"
				return cfg
			},
			giveExplicit: nil,
			wantKnow:     true,
			wantMem:      true,
			wantRetr:     true,
			wantAutoKnow: true,
			wantAutoMem:  true,
			wantAutoRetr: true,
		},
		{
			give: "nil explicitKeys + no dbPath → no auto-enable",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Session.DatabasePath = "" // clear default
				return cfg
			},
			giveExplicit: nil,
			wantKnow:     false,
			wantMem:      false,
			wantRetr:     false,
			wantAutoKnow: false,
			wantAutoMem:  false,
			wantAutoRetr: false,
		},
		{
			give: "knowledge explicitly disabled + dbPath → no auto-enable for knowledge",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Session.DatabasePath = "/tmp/test.db"
				cfg.Knowledge.Enabled = false
				return cfg
			},
			giveExplicit: map[string]bool{"knowledge.enabled": true},
			wantKnow:     false,
			wantMem:      true,
			wantRetr:     false, // retrieval follows knowledge
			wantAutoKnow: false,
			wantAutoMem:  true,
			wantAutoRetr: false,
		},
		{
			give: "knowledge already true + dbPath → no re-auto-enable",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Session.DatabasePath = "/tmp/test.db"
				cfg.Knowledge.Enabled = true
				return cfg
			},
			giveExplicit: map[string]bool{"knowledge.enabled": true},
			wantKnow:     true,
			wantMem:      true,
			wantRetr:     true,
			wantAutoKnow: false, // was already true, not auto-enabled
			wantAutoMem:  true,
			wantAutoRetr: true,
		},
		{
			give: "retrieval explicitly disabled → memory still auto-enabled",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Session.DatabasePath = "/tmp/test.db"
				cfg.Retrieval.Enabled = false
				return cfg
			},
			giveExplicit: map[string]bool{"retrieval.enabled": true},
			wantKnow:     true,
			wantMem:      true,
			wantRetr:     false,
			wantAutoKnow: true,
			wantAutoMem:  true,
			wantAutoRetr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			cfg := tt.giveCfg()
			set := ResolveContextAutoEnable(cfg, tt.giveExplicit)

			assert.Equal(t, tt.wantKnow, cfg.Knowledge.Enabled, "Knowledge.Enabled")
			assert.Equal(t, tt.wantMem, cfg.ObservationalMemory.Enabled, "Memory.Enabled")
			assert.Equal(t, tt.wantRetr, cfg.Retrieval.Enabled, "Retrieval.Enabled")
			assert.Equal(t, tt.wantAutoKnow, set.Knowledge, "AutoEnabled.Knowledge")
			assert.Equal(t, tt.wantAutoMem, set.Memory, "AutoEnabled.Memory")
			assert.Equal(t, tt.wantAutoRetr, set.Retrieval, "AutoEnabled.Retrieval")
		})
	}
}

func TestProbeEmbeddingProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		giveCfg   func() *Config
		wantProbe string
	}{
		{
			give: "already configured → returns existing",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Embedding.Provider = "my-openai"
				return cfg
			},
			wantProbe: "my-openai",
		},
		{
			give: "single local provider → auto-select",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Providers = map[string]ProviderConfig{
					"local-ollama": {Type: types.ProviderOllama},
				}
				return cfg
			},
			wantProbe: "local-ollama",
		},
		{
			give: "single remote provider → auto-select",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Providers = map[string]ProviderConfig{
					"my-openai": {Type: types.ProviderOpenAI, APIKey: "sk-test"},
				}
				return cfg
			},
			wantProbe: "my-openai",
		},
		{
			give: "local + remote → prefer local",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Providers = map[string]ProviderConfig{
					"my-openai":   {Type: types.ProviderOpenAI},
					"local-llama": {Type: types.ProviderOllama},
				}
				return cfg
			},
			wantProbe: "local-llama",
		},
		{
			give: "multiple remote → no auto-select (cost surprise prevention)",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Providers = map[string]ProviderConfig{
					"my-openai": {Type: types.ProviderOpenAI},
					"my-gemini": {Type: types.ProviderGemini},
				}
				return cfg
			},
			wantProbe: "",
		},
		{
			give: "only anthropic (no embedding support) → no auto-select",
			giveCfg: func() *Config {
				cfg := DefaultConfig()
				cfg.Providers = map[string]ProviderConfig{
					"my-claude": {Type: types.ProviderAnthropic},
				}
				return cfg
			},
			wantProbe: "",
		},
		{
			give: "no providers → no auto-select",
			giveCfg: func() *Config {
				return DefaultConfig()
			},
			wantProbe: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			cfg := tt.giveCfg()
			got := cfg.ProbeEmbeddingProvider()
			assert.Equal(t, tt.wantProbe, got)
		})
	}
}

func TestCollectExplicitKeys_Integration(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test.json")
	content := `{"knowledge":{"enabled":true},"retrieval":{"enabled":false}}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	keys := collectExplicitKeys(cfgPath, contextRelatedKeys)
	assert.True(t, keys["knowledge.enabled"])
	assert.True(t, keys["retrieval.enabled"])
	assert.False(t, keys["observationalMemory.enabled"])
	assert.False(t, keys["embedding.provider"])
}

func TestPresetExplicitKeys(t *testing.T) {
	t.Parallel()

	keys := PresetExplicitKeys("researcher")
	assert.True(t, keys["knowledge.enabled"])
	assert.True(t, keys["embedding.provider"])
	assert.True(t, keys["librarian.enabled"])

	keys = PresetExplicitKeys("minimal")
	assert.Empty(t, keys)
}
