package app

import (
	"testing"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/config"
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
		give     string
		wantGt0  bool
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

func TestInitSecurity_UnknownProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "nonexistent"

	_, _, _, err := initSecurity(cfg, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown security provider")
}

func TestInitSecurity_EnclaveNotImplemented(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "enclave"

	_, _, _, err := initSecurity(cfg, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestInitSecurity_LocalRequiresBootstrap(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "local"

	_, _, _, err := initSecurity(cfg, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires bootstrap")
}

func TestInitSecurity_KMSRequiresBootstrap(t *testing.T) {
	tests := []struct {
		give string
	}{
		{give: "aws-kms"},
		{give: "gcp-kms"},
		{give: "azure-kv"},
		{give: "pkcs11"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Security.Signer.Provider = tt.give

			_, _, _, err := initSecurity(cfg, nil, nil)
			require.Error(t, err)
			// Either "requires bootstrap" or KMS provider init error.
			assert.Error(t, err)
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
