package app

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/require"
)

func TestInitSupervisor(t *testing.T) {
	t.Skip("requires provider credentials")

	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	sv, err := initSupervisor(cfg)
	require.NoError(t, err)
	require.NotNil(t, sv, "expected supervisor to be initialized")
}
