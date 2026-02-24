//go:build kms_azure || kms_all

package security

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureKVProvider_ResolveKey(t *testing.T) {
	p := &AzureKVProvider{
		defaultKeyName: "my-default-key",
	}

	tests := []struct {
		give string
		want string
	}{
		{give: "local", want: "my-default-key"},
		{give: "default", want: "my-default-key"},
		{give: "", want: "my-default-key"},
		{give: "custom-key", want: "custom-key"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := p.resolveKey(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAzureKVProvider_NewWithoutVaultURL(t *testing.T) {
	cfg := config.KMSConfig{
		KeyID: "test-key",
		Azure: config.AzureKVConfig{
			VaultURL: "",
		},
	}

	_, err := newAzureKVProvider(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault URL is required")
}

func TestAzureKVProvider_NewWithoutKeyID(t *testing.T) {
	cfg := config.KMSConfig{
		KeyID: "",
		Azure: config.AzureKVConfig{
			VaultURL: "https://myvault.vault.azure.net",
		},
	}

	_, err := newAzureKVProvider(cfg)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrKMSInvalidKey)
}
