package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/orchestration"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- buildAgentOptions ---

func TestBuildAgentOptions_Defaults(t *testing.T) {
	cfg := config.DefaultConfig()

	opts := buildAgentOptions(cfg, nil)
	// Should always include token budget.
	require.NotEmpty(t, opts)
}

func TestBuildAgentOptions_ExplicitMaxTurns(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.MaxTurns = 25

	opts := buildAgentOptions(cfg, nil)
	// Should include token budget + max turns = at least 2 options.
	assert.GreaterOrEqual(t, len(opts), 2)
}

func TestBuildAgentOptions_MultiAgentDefaultMaxTurns(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.MultiAgent = true

	opts := buildAgentOptions(cfg, nil)
	// Should include token budget + default multi-agent max turns (50).
	assert.GreaterOrEqual(t, len(opts), 2)
}

func TestBuildAgentOptions_ErrorCorrectionDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	disabled := false
	cfg.Agent.ErrorCorrectionEnabled = &disabled

	opts := buildAgentOptions(cfg, nil)
	// With error correction disabled and nil kc, should only have token budget.
	assert.Len(t, opts, 1)
}

func TestBuildAgentOptions_ErrorCorrectionWithNilKC(t *testing.T) {
	cfg := config.DefaultConfig()
	// Error correction enabled (default) but no knowledge components.
	opts := buildAgentOptions(cfg, nil)
	// Should not add error correction option without knowledge components.
	assert.Len(t, opts, 1)
}

// --- ModelTokenBudget ---

func TestModelTokenBudget(t *testing.T) {
	tests := []struct {
		give    string
		wantGt0 bool
	}{
		{give: "gpt-4", wantGt0: true},
		{give: "gemini-2.0-flash", wantGt0: true},
		{give: "claude-3-opus-20240229", wantGt0: true},
		{give: "unknown-model", wantGt0: true}, // should return a default
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			budget := adk.ModelTokenBudget(tt.give)
			if tt.wantGt0 {
				assert.Greater(t, budget, 0, "expected positive token budget for model %q", tt.give)
			}
		})
	}
}

// --- initSecurity branching ---

func TestInitSecurity_EmptyProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = ""

	crypto, keys, secrets, err := initSecurity(cfg, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, crypto)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)
}

func TestInitSecurity_UnsupportedProvider(t *testing.T) {
	tests := []struct {
		give string
	}{
		{give: "enclave"},
		{give: "nonexistent"},
		{give: "hashicorp-vault"},
		{give: ""},
	}

	validProviders := []string{"local", "rpc", "aws-kms", "gcp-kms", "azure-kv", "pkcs11"}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Security.Signer.Provider = tt.give

			_, _, _, err := initSecurity(cfg, nil, nil)

			if tt.give == "" {
				// Empty provider is a no-op, not an error.
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported security provider")
			assert.Contains(t, err.Error(), tt.give)

			// Verify the error message lists all valid providers.
			for _, valid := range validProviders {
				assert.Contains(t, err.Error(), valid,
					"error should list valid provider %q", valid)
			}
		})
	}
}

func TestInitSecurity_LocalRequiresBootstrap(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "local"

	_, _, _, err := initSecurity(cfg, nil, nil)
	require.Error(t, err)
	assert.EqualError(t, err, "local security provider requires bootstrap")
}

func TestInitSecurity_KMSRequiresBootstrap(t *testing.T) {
	tests := []struct {
		give        string
		wantTagHint string
	}{
		{give: "aws-kms", wantTagHint: "kms_aws"},
		{give: "gcp-kms", wantTagHint: "kms_gcp"},
		{give: "azure-kv", wantTagHint: "kms_azure"},
		{give: "pkcs11", wantTagHint: "kms_pkcs11"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Security.Signer.Provider = tt.give

			_, _, _, err := initSecurity(cfg, nil, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.give)
			assert.Contains(t, err.Error(), "support not compiled")
			assert.Contains(t, err.Error(), "rebuild with -tags "+tt.wantTagHint)
		})
	}
}

// --- initAuth ---

func TestInitAuth_NoProviders(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Auth.Providers = nil

	auth := initAuth(cfg, nil)
	assert.Nil(t, auth, "expected nil auth when no providers configured")
}

func TestInitAuth_EmptyProviders(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Auth.Providers = map[string]config.OIDCProviderConfig{}

	auth := initAuth(cfg, nil)
	assert.Nil(t, auth, "expected nil auth when providers map is empty")
}

func TestBuildPromptBuilder_LoadsConfiguredPromptDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("Custom identity"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CUSTOM.md"), []byte("Custom extension"), 0o644))

	builder := buildPromptBuilder(&config.AgentConfig{PromptsDir: dir})
	prompt := builder.Build()

	assert.Contains(t, prompt, "Custom identity")
	assert.Contains(t, prompt, "Custom extension")
}

func TestBuildSubAgentPromptFunc_InjectsIdentityAndAgentOverrides(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("Global identity must not leak"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SAFETY.md"), []byte("Shared safety"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "TOOL_USAGE.md"), []byte("Global tool usage must not leak"), 0o644))
	agentDir := filepath.Join(dir, "agents", "reviewer")
	require.NoError(t, os.MkdirAll(agentDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(agentDir, "IDENTITY.md"), []byte("Reviewer override"), 0o644))

	buildPrompt := buildSubAgentPromptFunc(&config.AgentConfig{PromptsDir: dir})
	got := buildPrompt("reviewer", "Default reviewer identity")

	assert.Contains(t, got, "Reviewer override")
	assert.Contains(t, got, "Shared safety")
	assert.NotContains(t, got, "Global identity must not leak")
	assert.NotContains(t, got, "Global tool usage must not leak")
	assert.NotContains(t, got, "Default reviewer identity")
}

func TestIsolatedAgentNames_UsesDefaultsOrCustomSpecs(t *testing.T) {
	t.Parallel()

	defaultNames := isolatedAgentNames(nil)
	assert.NotEmpty(t, defaultNames, "default orchestration specs should include isolated agents")

	customNames := isolatedAgentNames([]orchestration.AgentSpec{
		{Name: "shared"},
		{Name: "isolated-one", SessionIsolation: true},
		{Name: "isolated-two", SessionIsolation: true},
	})
	assert.Equal(t, []string{"isolated-one", "isolated-two"}, customNames)
}

func TestCatalogSourceAdapter_BuildsModeFilteredCatalogSection(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "knowledge",
		Description: "Knowledge tools",
		Enabled:     true,
	})
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "disabled",
		Description: "Disabled tools",
		ConfigKey:   "tools.disabled",
		Enabled:     false,
	})
	catalog.Register("knowledge", []*agent.Tool{
		{Name: "save_knowledge", Description: "save", SafetyLevel: agent.SafetyLevelModerate},
		{Name: "search_knowledge", Description: "search", SafetyLevel: agent.SafetyLevelSafe},
		{
			Name:        "deferred_knowledge",
			Description: "deferred",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability:  agent.ToolCapability{Exposure: agent.ExposureDeferred},
		},
	})
	catalog.Register("disabled", []*agent.Tool{
		{Name: "disabled_tool", Description: "disabled", SafetyLevel: agent.SafetyLevelSafe},
	})
	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"narrow": {
			Name:  "narrow",
			Tools: []string{"search_knowledge", "deferred_knowledge", "missing_tool"},
		},
	}
	adapter := &catalogSourceAdapter{catalog: catalog, cfg: cfg}

	modeSection := adapter.BuildToolCatalogSection("narrow")
	assert.Contains(t, modeSection, "## Tools Available in `narrow` Mode")
	assert.Contains(t, modeSection, "search_knowledge")
	assert.NotContains(t, modeSection, "save_knowledge")
	assert.NotContains(t, modeSection, "deferred_knowledge")
	assert.Contains(t, modeSection, "Only tools in this mode's allowlist")

	defaultSection := adapter.BuildToolCatalogSection("")
	assert.Contains(t, defaultSection, "save_knowledge")
	assert.Contains(t, defaultSection, "search_knowledge")
	assert.Contains(t, defaultSection, "Disabled categories (enable via config): disabled (tools.disabled)")
	assert.Contains(t, defaultSection, "Additional 1 specialized tools available via builtin_search.")
}

func TestCatalogSourceAdapter_TruncatesLongCategoryToolList(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "bulk",
		Description: "Bulk tools",
		Enabled:     true,
	})
	var tools []*agent.Tool
	for i := 0; i < 10; i++ {
		tools = append(tools, &agent.Tool{
			Name:        "tool_" + string(rune('a'+i)),
			Description: "tool",
			SafetyLevel: agent.SafetyLevelSafe,
		})
	}
	catalog.Register("bulk", tools)

	section := (&catalogSourceAdapter{catalog: catalog}).BuildToolCatalogSection("")

	assert.Contains(t, section, "... +2 more")
	for _, name := range []string{"tool_a", "tool_b", "tool_c", "tool_d", "tool_e", "tool_f", "tool_g", "tool_h"} {
		assert.Contains(t, section, name)
	}
	assert.False(t, strings.Contains(section, "tool_i"), "catalog should hide the ninth tool")
	assert.False(t, strings.Contains(section, "tool_j"), "catalog should hide the tenth tool")
}

func TestModeResolverAdapter_LookupModeHint(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"focused": {
			Name:       "focused",
			SystemHint: "Stay within focused mode.",
		},
	}
	adapter := &modeResolverAdapter{cfg: cfg}

	assert.Equal(t, "Stay within focused mode.", adapter.LookupModeHint("focused"))
	assert.Empty(t, adapter.LookupModeHint(""))
	assert.Empty(t, adapter.LookupModeHint("missing"))
	assert.Empty(t, (*modeResolverAdapter)(nil).LookupModeHint("focused"))
}
