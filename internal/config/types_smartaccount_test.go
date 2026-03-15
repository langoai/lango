package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartAccountConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		giveCfg SmartAccountConfig
		wantErr string
	}{
		{
			give:    "disabled config is always valid",
			giveCfg: SmartAccountConfig{Enabled: false},
		},
		{
			give: "all required fields present",
			giveCfg: SmartAccountConfig{
				Enabled:           true,
				EntryPointAddress: "0x1234",
				FactoryAddress:    "0x5678",
				BundlerURL:        "https://bundler.example.com",
			},
		},
		{
			give: "missing entryPointAddress",
			giveCfg: SmartAccountConfig{
				Enabled:        true,
				FactoryAddress: "0x5678",
				BundlerURL:     "https://bundler.example.com",
			},
			wantErr: "smartAccount.entryPointAddress is required",
		},
		{
			give: "missing factoryAddress",
			giveCfg: SmartAccountConfig{
				Enabled:           true,
				EntryPointAddress: "0x1234",
				BundlerURL:        "https://bundler.example.com",
			},
			wantErr: "smartAccount.factoryAddress is required",
		},
		{
			give: "missing bundlerURL",
			giveCfg: SmartAccountConfig{
				Enabled:           true,
				EntryPointAddress: "0x1234",
				FactoryAddress:    "0x5678",
			},
			wantErr: "smartAccount.bundlerURL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			err := tt.giveCfg.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
