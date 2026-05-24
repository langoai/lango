package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPromptBuilderDefaultReturnsContent(t *testing.T) {
	t.Parallel()

	builder := buildPromptBuilder(&config.AgentConfig{})

	require.NotNil(t, builder)
	assert.NotEmpty(t, builder.Build())
}

func TestBuildSubAgentPromptFuncCustomDirWithoutSafetyKeepsDefaultOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("Global identity must not leak"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "TOOL_USAGE.md"), []byte("Global tool usage must not leak"), 0o644))

	buildPrompt := buildSubAgentPromptFunc(&config.AgentConfig{PromptsDir: dir})
	got := buildPrompt("reviewer", "Default reviewer identity")

	assert.Contains(t, got, "Default reviewer identity")
	assert.NotContains(t, got, "Global identity must not leak")
	assert.NotContains(t, got, "Global tool usage must not leak")
}

func TestCatalogSourceAdapterModeWithEmptyAllowlistFallsBackToVisibleTools(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "core",
		Description: "Core tools",
		Enabled:     true,
	})
	catalog.Register("core", []*agent.Tool{
		{Name: "visible_tool", Description: "visible", SafetyLevel: agent.SafetyLevelSafe},
	})

	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"empty-allowlist": {
			Name:  "empty-allowlist",
			Tools: []string{},
		},
	}
	adapter := &catalogSourceAdapter{catalog: catalog, cfg: cfg}

	section := adapter.BuildToolCatalogSection("empty-allowlist")

	assert.Contains(t, section, "## Tools Available in `empty-allowlist` Mode")
	assert.Contains(t, section, "visible_tool")
	assert.Contains(t, section, "Only tools in this mode's allowlist")
}

func TestCatalogSourceAdapterDisabledCategoryWithoutConfigKeyIsNotReported(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "enabled",
		Description: "Enabled tools",
		Enabled:     true,
	})
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "disabled_without_key",
		Description: "Disabled without config key",
		Enabled:     false,
	})
	catalog.Register("enabled", []*agent.Tool{
		{Name: "enabled_tool", Description: "enabled", SafetyLevel: agent.SafetyLevelSafe},
	})

	section := (&catalogSourceAdapter{catalog: catalog}).BuildToolCatalogSection("")

	assert.Contains(t, section, "## Available Tool Categories")
	assert.Contains(t, section, "enabled_tool")
	assert.NotContains(t, section, "Disabled categories (enable via config):")
	assert.NotContains(t, section, "disabled_without_key")
}

func TestInitAuthWithLocalOIDCProviderReturnsManager(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{
				"issuer": %q,
				"authorization_endpoint": %q,
				"token_endpoint": %q,
				"jwks_uri": %q,
				"response_types_supported": ["code"],
				"subject_types_supported": ["public"],
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, serverURL(r), serverURL(r)+"/auth", serverURL(r)+"/token", serverURL(r)+"/keys")
		case "/keys":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"keys":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := config.DefaultConfig()
	cfg.Auth.Providers = map[string]config.OIDCProviderConfig{
		"local": {
			IssuerURL:    server.URL,
			ClientID:     "client",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
			Scopes:       []string{"openid"},
		},
	}

	auth := initAuth(cfg, &stubSessionStore{})

	assert.NotNil(t, auth)
}

func serverURL(r *http.Request) string {
	return "http://" + r.Host
}
