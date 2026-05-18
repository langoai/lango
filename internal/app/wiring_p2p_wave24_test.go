package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestWave24InitSecurityCoversDisabledAndEarlyErrorBranches(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = ""

	provider, keys, secrets, err := initSecurity(cfg, &stubSessionStore{}, nil)
	require.NoError(t, err)
	assert.Nil(t, provider)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)

	cfg.Security.Signer.Provider = "local"
	provider, keys, secrets, err = initSecurity(cfg, &stubSessionStore{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "local security provider requires bootstrap")
	assert.Nil(t, provider)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)

	cfg.Security.Signer.Provider = "rpc"
	provider, keys, secrets, err = initSecurity(cfg, &stubSessionStore{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rpc security provider requires EntStore")
	assert.Nil(t, provider)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)

	cfg.Security.Signer.Provider = "unsupported"
	provider, keys, secrets, err = initSecurity(cfg, &stubSessionStore{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unsupported security provider "unsupported"`)
	assert.Nil(t, provider)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)
}

func TestWave24InitAuthReturnsNilForMissingOrInvalidProviders(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	assert.Nil(t, initAuth(cfg, &stubSessionStore{}))

	cfg.Auth.Providers = map[string]config.OIDCProviderConfig{
		"bad": {IssuerURL: "://not-a-valid-issuer-url"},
	}
	assert.Nil(t, initAuth(cfg, &stubSessionStore{}))
}

func TestWave24BuildSubAgentPromptFuncStripsGlobalSectionsAndKeepsSharedPrompt(t *testing.T) {
	t.Parallel()

	promptsDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "AGENTS.md"),
		[]byte("GLOBAL IDENTITY MUST NOT LEAK"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "TOOL_USAGE.md"),
		[]byte("GLOBAL TOOL RULES MUST NOT LEAK"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "SAFETY.md"),
		[]byte("Shared wave24 safety."), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "CONVERSATION_RULES.md"),
		[]byte("Shared wave24 conversation rules."), 0o600))

	buildPrompt := buildSubAgentPromptFunc(&config.AgentConfig{PromptsDir: promptsDir})
	got := buildPrompt("reviewer", "Default reviewer instruction.")

	assert.Contains(t, got, "Default reviewer instruction.")
	assert.Contains(t, got, "Shared wave24 safety.")
	assert.Contains(t, got, "Shared wave24 conversation rules.")
	assert.NotContains(t, got, "GLOBAL IDENTITY MUST NOT LEAK")
	assert.NotContains(t, got, "GLOBAL TOOL RULES MUST NOT LEAK")
}

func TestWave24InitP2PReturnsNilWhenNodeCreationFailsOnInvalidListenAddr(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "keys")
	cfg.P2P.ListenAddrs = []string{"not-a-multiaddr"}

	got := initP2P(cfg, &wiringP2PWallet{}, nil, nil, nil, nil, nil, nil, "")

	assert.Nil(t, got)
}
