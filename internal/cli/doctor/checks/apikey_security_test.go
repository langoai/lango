package checks

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/config"
)

func TestAPIKeySecurityCheck_Run(t *testing.T) {
	tests := []struct {
		give       string
		wantStatus Status
		cfg        *config.Config
	}{
		{
			give:       "nil config",
			wantStatus: StatusSkip,
			cfg:        nil,
		},
		{
			give:       "env var reference key",
			wantStatus: StatusPass,
			cfg: &config.Config{
				Providers: map[string]config.ProviderConfig{
					"openai": {Type: "openai", APIKey: "${OPENAI_API_KEY}"},
				},
			},
		},
		{
			give:       "plaintext key",
			wantStatus: StatusWarn,
			cfg: &config.Config{
				Providers: map[string]config.ProviderConfig{
					"openai": {Type: "openai", APIKey: "sk-plaintext-key-12345"},
				},
			},
		},
		{
			give:       "mixed keys",
			wantStatus: StatusWarn,
			cfg: &config.Config{
				Providers: map[string]config.ProviderConfig{
					"openai":    {Type: "openai", APIKey: "${OPENAI_API_KEY}"},
					"anthropic": {Type: "anthropic", APIKey: "sk-ant-plaintext"},
				},
			},
		},
		{
			give:       "no providers configured",
			wantStatus: StatusSkip,
			cfg: &config.Config{
				Providers: map[string]config.ProviderConfig{},
			},
		},
		{
			give:       "providers with empty keys",
			wantStatus: StatusSkip,
			cfg: &config.Config{
				Providers: map[string]config.ProviderConfig{
					"openai": {Type: "openai", APIKey: ""},
				},
			},
		},
	}

	check := &APIKeySecurityCheck{}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := check.Run(context.Background(), tt.cfg)

			if result.Status != tt.wantStatus {
				t.Errorf("expected %v, got %v: %s", tt.wantStatus, result.Status, result.Message)
			}
		})
	}
}
