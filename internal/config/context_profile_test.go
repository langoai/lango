package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyContextProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give            ContextProfileName
		explicitKeys    map[string]bool
		wantKnowledge   bool
		wantMemory      bool
		wantLibrarian   bool
		wantGraph       bool
	}{
		{
			give:          ContextProfileOff,
			wantKnowledge: false, wantMemory: false, wantLibrarian: false, wantGraph: false,
		},
		{
			give:          ContextProfileLite,
			wantKnowledge: true, wantMemory: true, wantLibrarian: false, wantGraph: false,
		},
		{
			give:          ContextProfileBalanced,
			wantKnowledge: true, wantMemory: true, wantLibrarian: true, wantGraph: false,
		},
		{
			give:          ContextProfileFull,
			wantKnowledge: true, wantMemory: true, wantLibrarian: true, wantGraph: true,
		},
		{
			give:          "",
			wantKnowledge: false, wantMemory: false, wantLibrarian: false, wantGraph: false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			t.Parallel()
			cfg := &Config{}
			cfg.ContextProfile = tt.give
			ApplyContextProfile(cfg, tt.explicitKeys)
			assert.Equal(t, tt.wantKnowledge, cfg.Knowledge.Enabled, "Knowledge")
			assert.Equal(t, tt.wantMemory, cfg.ObservationalMemory.Enabled, "Memory")
			assert.Equal(t, tt.wantLibrarian, cfg.Librarian.Enabled, "Librarian")
			assert.Equal(t, tt.wantGraph, cfg.Graph.Enabled, "Graph")
		})
	}
}

func TestApplyContextProfile_ExplicitOverrides(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		profile      ContextProfileName
		explicitKeys map[string]bool
		preSet       func(*Config)
		wantField    string
		wantValue    bool
	}{
		{
			give:         "explicit false preserved against balanced",
			profile:      ContextProfileBalanced,
			explicitKeys: map[string]bool{"knowledge.enabled": true},
			preSet:       func(c *Config) { c.Knowledge.Enabled = false },
			wantField:    "Knowledge",
			wantValue:    false,
		},
		{
			give:         "explicit true preserved against off",
			profile:      ContextProfileOff,
			explicitKeys: map[string]bool{"graph.enabled": true},
			preSet:       func(c *Config) { c.Graph.Enabled = true },
			wantField:    "Graph",
			wantValue:    true,
		},
		{
			give:         "nil explicitKeys means no overrides",
			profile:      ContextProfileBalanced,
			explicitKeys: nil,
			wantField:    "Knowledge",
			wantValue:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{ContextProfile: tt.profile}
			if tt.preSet != nil {
				tt.preSet(cfg)
			}
			ApplyContextProfile(cfg, tt.explicitKeys)
			switch tt.wantField {
			case "Knowledge":
				assert.Equal(t, tt.wantValue, cfg.Knowledge.Enabled)
			case "Graph":
				assert.Equal(t, tt.wantValue, cfg.Graph.Enabled)
			}
		})
	}
}

func TestApplyContextProfile_InvalidProfile(t *testing.T) {
	t.Parallel()
	cfg := &Config{ContextProfile: "turbo"}
	ApplyContextProfile(cfg, nil)
	// Invalid profile is a no-op; fields unchanged from zero values.
	assert.False(t, cfg.Knowledge.Enabled)
}

func TestCollectExplicitKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test.json")
	content := `{"knowledge":{"enabled":false},"graph":{"enabled":true}}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	keys := collectExplicitKeys(cfgPath, contextRelatedKeys)
	assert.True(t, keys["knowledge.enabled"], "knowledge.enabled should be explicit")
	assert.True(t, keys["graph.enabled"], "graph.enabled should be explicit")
	assert.False(t, keys["librarian.enabled"], "librarian.enabled should not be explicit")
	assert.False(t, keys["observationalMemory.enabled"], "observationalMemory.enabled should not be explicit")
}

func TestCollectExplicitKeys_NoFile(t *testing.T) {
	t.Parallel()
	keys := collectExplicitKeys("", contextRelatedKeys)
	assert.Nil(t, keys)
}

func TestCollectExplicitKeys_NonexistentFile(t *testing.T) {
	t.Parallel()
	keys := collectExplicitKeys("/nonexistent/path.json", contextRelatedKeys)
	assert.Nil(t, keys)
}

func TestLoad_ContextProfile_Balanced(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")
	content := `{"contextProfile":"balanced","logging":{"level":"info","format":"console"}}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	result, err := Load(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, ContextProfileBalanced, result.Config.ContextProfile)
	assert.True(t, result.Config.Knowledge.Enabled)
	assert.True(t, result.Config.ObservationalMemory.Enabled)
	assert.True(t, result.Config.Librarian.Enabled)
	assert.False(t, result.Config.Graph.Enabled)
}

func TestLoad_ContextProfile_InvalidRejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")
	content := `{"contextProfile":"turbo","logging":{"level":"info","format":"console"}}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	result, err := Load(cfgPath)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid contextProfile")
}

func TestLoad_ExplicitOverridePreserved(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")
	content := `{"contextProfile":"balanced","knowledge":{"enabled":false},"logging":{"level":"info","format":"console"}}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	result, err := Load(cfgPath)
	require.NoError(t, err)

	assert.False(t, result.Config.Knowledge.Enabled, "explicit false should be preserved")
	assert.True(t, result.Config.ObservationalMemory.Enabled, "non-explicit should be set by profile")
	assert.True(t, result.ExplicitKeys["knowledge.enabled"])
}
