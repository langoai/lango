package smartaccount

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/module"
)

func TestInitSmartAccountDepsRejectsInvalidPrerequisites(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*config.Config)
		wantErr string
	}{
		{
			name: "payment disabled",
			mutate: func(cfg *config.Config) {
				cfg.SmartAccount.Enabled = true
				cfg.Payment.Enabled = false
			},
			wantErr: "smart account requires payment to be enabled",
		},
		{
			name: "invalid smart account config",
			mutate: func(cfg *config.Config) {
				cfg.SmartAccount.Enabled = true
				cfg.Payment.Enabled = true
			},
			wantErr: "smart account config: smartAccount.entryPointAddress is required",
		},
		{
			name: "storage unavailable",
			mutate: func(cfg *config.Config) {
				cfg.SmartAccount.Enabled = true
				cfg.Payment.Enabled = true
				cfg.SmartAccount.EntryPointAddress = "0x0000000000000000000000000000000000000001"
				cfg.SmartAccount.FactoryAddress = "0x0000000000000000000000000000000000000002"
				cfg.SmartAccount.BundlerURL = "http://127.0.0.1:4337"
			},
			wantErr: "smart account storage unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			tt.mutate(cfg)

			deps, err := initSmartAccountDeps(&bootstrap.Result{Config: cfg})

			require.Error(t, err)
			assert.Nil(t, deps)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestInitPaymasterProviderBranches(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.SmartAccountPaymasterConfig
		wantNil  bool
		wantType string
	}{
		{
			name: "permit circle without dependencies",
			cfg: config.SmartAccountPaymasterConfig{
				Mode:     "permit",
				Provider: "circle",
			},
			wantNil: true,
		},
		{
			name:    "rpc mode without url",
			cfg:     config.SmartAccountPaymasterConfig{Provider: "circle"},
			wantNil: true,
		},
		{
			name: "unknown rpc provider",
			cfg: config.SmartAccountPaymasterConfig{
				Provider: "unknown",
				RPCURL:   "http://127.0.0.1:4337",
			},
			wantNil: true,
		},
		{
			name: "circle rpc provider",
			cfg: config.SmartAccountPaymasterConfig{
				Provider: "circle",
				RPCURL:   "http://127.0.0.1:4337",
			},
			wantType: "circle",
		},
		{
			name: "pimlico rpc provider",
			cfg: config.SmartAccountPaymasterConfig{
				Provider: "pimlico",
				RPCURL:   "http://127.0.0.1:4337",
				PolicyID: "policy-1",
			},
			wantType: "pimlico",
		},
		{
			name: "alchemy rpc provider",
			cfg: config.SmartAccountPaymasterConfig{
				Provider: "alchemy",
				RPCURL:   "http://127.0.0.1:4337",
				PolicyID: "policy-1",
			},
			wantType: "alchemy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := initPaymasterProvider(tt.cfg, nil, nil, 8453)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.wantType, got.Type())
		})
	}
}

func TestRegisterDefaultModulesRegistersConfiguredDescriptors(t *testing.T) {
	reg := module.NewRegistry()
	cfg := config.SmartAccountModulesConfig{
		SessionValidatorAddress: "0x0000000000000000000000000000000000000011",
		SpendingHookAddress:     "0x0000000000000000000000000000000000000022",
		EscrowExecutorAddress:   "0x0000000000000000000000000000000000000033",
	}

	registerDefaultModules(reg, cfg)

	validator, err := reg.Get(common.HexToAddress(cfg.SessionValidatorAddress))
	require.NoError(t, err)
	assert.Equal(t, "LangoSessionValidator", validator.Name)
	assert.Equal(t, sa.ModuleTypeValidator, validator.Type)

	hook, err := reg.Get(common.HexToAddress(cfg.SpendingHookAddress))
	require.NoError(t, err)
	assert.Equal(t, "LangoSpendingHook", hook.Name)
	assert.Equal(t, sa.ModuleTypeHook, hook.Type)

	executor, err := reg.Get(common.HexToAddress(cfg.EscrowExecutorAddress))
	require.NoError(t, err)
	assert.Equal(t, "LangoEscrowExecutor", executor.Name)
	assert.Equal(t, sa.ModuleTypeExecutor, executor.Type)
}

func TestRegisterDefaultModulesSkipsEmptyConfig(t *testing.T) {
	reg := module.NewRegistry()

	registerDefaultModules(reg, config.SmartAccountModulesConfig{})

	assert.Empty(t, reg.List())
}
