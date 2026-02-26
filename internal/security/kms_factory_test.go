package security

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestKMSProviderName_Valid(t *testing.T) {
	tests := []struct {
		name  KMSProviderName
		valid bool
	}{
		{KMSProviderAWS, true},
		{KMSProviderGCP, true},
		{KMSProviderAzure, true},
		{KMSProviderPKCS11, true},
		{"unknown", false},
		{"", false},
		{"local", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.name.Valid())
		})
	}
}

func TestKMSProviderName_Constants(t *testing.T) {
	assert.Equal(t, KMSProviderName("aws-kms"), KMSProviderAWS)
	assert.Equal(t, KMSProviderName("gcp-kms"), KMSProviderGCP)
	assert.Equal(t, KMSProviderName("azure-kv"), KMSProviderAzure)
	assert.Equal(t, KMSProviderName("pkcs11"), KMSProviderPKCS11)
}

func TestNewKMSProvider_UnknownProvider(t *testing.T) {
	provider, err := NewKMSProvider("unknown-provider", config.KMSConfig{})
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unknown KMS provider")
	assert.Contains(t, err.Error(), "unknown-provider")
}
