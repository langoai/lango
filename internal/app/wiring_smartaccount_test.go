package app

import (
	"context"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// initSmartAccount
// ---------------------------------------------------------------------------

func TestInitSmartAccount_DisabledReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SmartAccount.Enabled = false

	result := initSmartAccount(cfg, nil, nil, nil, nil)

	assert.Nil(t, result, "expected nil when smart account is disabled")
}

func TestInitSmartAccount_NilPaymentComponentsReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SmartAccount.Enabled = true
	// Even with enabled config, nil payment components should cause early return.

	result := initSmartAccount(cfg, nil, nil, nil, nil)

	assert.Nil(t, result, "expected nil when payment components are nil")
}

func TestInitSmartAccount_IncompleteConfigReturnsNil(t *testing.T) {
	// Smart account is enabled and payment components exist, but required
	// config fields (entryPointAddress, factoryAddress, bundlerURL) are missing.
	cfg := config.DefaultConfig()
	cfg.SmartAccount.Enabled = true
	pc := &paymentComponents{}

	result := initSmartAccount(cfg, pc, nil, nil, nil)

	assert.Nil(t, result, "expected nil when config validation fails due to missing fields")
}

func TestInitSmartAccount_DisabledBranch_TableDriven(t *testing.T) {
	tests := []struct {
		give      string
		giveOn    bool
		givePC    *paymentComponents
		giveEconc *economyComponents
		giveBus   *eventbus.Bus
		wantNil   bool
	}{
		{
			give:    "disabled config returns nil",
			giveOn:  false,
			givePC:  nil,
			wantNil: true,
		},
		{
			give:    "enabled but nil payment returns nil",
			giveOn:  true,
			givePC:  nil,
			wantNil: true,
		},
		{
			give:    "disabled with non-nil payment returns nil",
			giveOn:  false,
			givePC:  &paymentComponents{},
			wantNil: true,
		},
		{
			give:      "disabled with all deps returns nil",
			giveOn:    false,
			givePC:    &paymentComponents{},
			giveEconc: &economyComponents{},
			giveBus:   eventbus.New(),
			wantNil:   true,
		},
		{
			give:      "enabled with payment but missing config fields returns nil",
			giveOn:    true,
			givePC:    &paymentComponents{},
			giveEconc: &economyComponents{},
			giveBus:   eventbus.New(),
			wantNil:   true, // Validate() fails: missing entryPointAddress, factoryAddress, bundlerURL
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.SmartAccount.Enabled = tt.giveOn

			result := initSmartAccount(cfg, tt.givePC, tt.giveEconc, tt.giveBus, nil)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// smartAccountComponents accessor methods
// ---------------------------------------------------------------------------

func TestSmartAccountComponents_AccessorsReturnNilOnZeroValue(t *testing.T) {
	sac := &smartAccountComponents{}

	assert.Nil(t, sac.SessionManager(), "SessionManager should be nil on zero value")
	assert.Nil(t, sac.PolicyEngine(), "PolicyEngine should be nil on zero value")
	assert.Nil(t, sac.OnChainTracker(), "OnChainTracker should be nil on zero value")
	assert.Nil(t, sac.PaymasterProvider(), "PaymasterProvider should be nil on zero value")
	assert.Nil(t, sac.ModuleRegistry(), "ModuleRegistry should be nil on zero value")
	assert.Nil(t, sac.BundlerClient(), "BundlerClient should be nil on zero value")
}

func TestToOnChainPolicy_ConvertsSessionPolicyFields(t *testing.T) {
	t.Parallel()

	target := common.HexToAddress("0xb00000000000000000000000000000000000000b")
	paymaster := common.HexToAddress("0xc00000000000000000000000000000000000000c")
	validAfter := int64(1_700_000_000)
	validUntil := int64(1_700_003_600)

	got := toOnChainPolicy(sa.SessionPolicy{
		AllowedTargets:    []common.Address{target},
		AllowedFunctions:  []string{"0xa9059cbb", "0xdeadbeef00", "0x12"},
		SpendLimit:        big.NewInt(123),
		SpentAmount:       big.NewInt(45),
		ValidAfter:        time.Unix(validAfter, 0),
		ValidUntil:        time.Unix(validUntil, 0),
		AllowedPaymasters: []common.Address{paymaster},
	})

	value := reflect.ValueOf(got)
	require.Equal(t, reflect.Struct, value.Kind())
	assert.Equal(t, []common.Address{target}, value.FieldByName("AllowedTargets").Interface())
	assert.Equal(t, big.NewInt(123), value.FieldByName("SpendLimit").Interface())
	assert.Equal(t, big.NewInt(45), value.FieldByName("SpentAmount").Interface())
	assert.Equal(t, big.NewInt(validAfter), value.FieldByName("ValidAfter").Interface())
	assert.Equal(t, big.NewInt(validUntil), value.FieldByName("ValidUntil").Interface())
	assert.Equal(t, true, value.FieldByName("Active").Interface())
	assert.Equal(t, []common.Address{paymaster}, value.FieldByName("AllowedPaymasters").Interface())

	selectors := value.FieldByName("AllowedFunctions").Interface().([][4]byte)
	require.Len(t, selectors, 2, "short selectors should be ignored")
	assert.Equal(t, [4]byte{0xa9, 0x05, 0x9c, 0xbb}, selectors[0])
	assert.Equal(t, [4]byte{0xde, 0xad, 0xbe, 0xef}, selectors[1])
}

func TestToOnChainPolicy_DefaultsNilSpendAmountsToZero(t *testing.T) {
	t.Parallel()

	got := toOnChainPolicy(sa.SessionPolicy{})
	value := reflect.ValueOf(got)

	assert.Equal(t, big.NewInt(0), value.FieldByName("SpendLimit").Interface())
	assert.Equal(t, big.NewInt(0), value.FieldByName("SpentAmount").Interface())
	assert.Equal(t, true, value.FieldByName("Active").Interface())
}

func TestRegisterDefaultModules_ConfiguredAddresses(t *testing.T) {
	t.Parallel()

	sessionValidator := common.HexToAddress("0xd00000000000000000000000000000000000000d")
	spendingHook := common.HexToAddress("0xe00000000000000000000000000000000000000e")
	escrowExecutor := common.HexToAddress("0xf00000000000000000000000000000000000000f")
	registry := module.NewRegistry()

	registerDefaultModules(registry, config.SmartAccountModulesConfig{
		SessionValidatorAddress: sessionValidator.Hex(),
		SpendingHookAddress:     spendingHook.Hex(),
		EscrowExecutorAddress:   escrowExecutor.Hex(),
	})

	require.Len(t, registry.List(), 3)
	assertRegisteredModule(t, registry, sessionValidator, "LangoSessionValidator", sa.ModuleTypeValidator)
	assertRegisteredModule(t, registry, spendingHook, "LangoSpendingHook", sa.ModuleTypeHook)
	assertRegisteredModule(t, registry, escrowExecutor, "LangoEscrowExecutor", sa.ModuleTypeExecutor)
	assert.Len(t, registry.ListByType(sa.ModuleTypeFallback), 0)
}

func TestInitPaymasterProvider_RPCModeBranches(t *testing.T) {
	t.Parallel()

	base := config.SmartAccountPaymasterConfig{Mode: "rpc"}

	assert.Nil(t, initPaymasterProvider(base, nil, nil, 84532))

	unknown := base
	unknown.RPCURL = "https://paymaster.invalid"
	unknown.Provider = "unknown"
	assert.Nil(t, initPaymasterProvider(unknown, nil, nil, 84532))

	for _, tc := range []struct {
		name     string
		provider string
		wantType string
	}{
		{name: "circle", provider: "circle", wantType: "circle+recovery"},
		{name: "pimlico", provider: "pimlico", wantType: "pimlico+recovery"},
		{name: "alchemy", provider: "alchemy", wantType: "alchemy+recovery"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := base
			cfg.RPCURL = "https://paymaster.invalid"
			cfg.Provider = tc.provider
			cfg.PolicyID = "policy-1"

			provider := initPaymasterProvider(cfg, nil, nil, 84532)
			require.NotNil(t, provider)
			assert.Equal(t, tc.wantType, provider.Type())
		})
	}
}

func TestInitPaymasterProvider_PermitModeBranches(t *testing.T) {
	t.Parallel()

	cfg := config.SmartAccountPaymasterConfig{
		Mode:             "permit",
		Provider:         "circle",
		PaymasterAddress: "0x1000000000000000000000000000000000000001",
		TokenAddress:     "0x2000000000000000000000000000000000000002",
		FallbackMode:     "direct",
	}

	assert.Nil(t, initPaymasterProvider(cfg, nil, fakeEthCaller{}, 84532))
	assert.Nil(t, initPaymasterProvider(cfg, fakeWalletProvider{}, nil, 84532))

	provider := initPaymasterProvider(cfg, fakeWalletProvider{}, fakeEthCaller{}, 84532)
	require.NotNil(t, provider)
	assert.Equal(t, "circle-permit+recovery", provider.Type())
}

func assertRegisteredModule(t *testing.T, registry *module.Registry, addr common.Address, name string, moduleType sa.ModuleType) {
	t.Helper()

	desc, err := registry.Get(addr)
	require.NoError(t, err)
	assert.Equal(t, name, desc.Name)
	assert.Equal(t, addr, desc.Address)
	assert.Equal(t, moduleType, desc.Type)
	assert.Equal(t, "1.0.0", desc.Version)
}

type fakeWalletProvider struct{}

func (fakeWalletProvider) SignTransaction(context.Context, []byte) ([]byte, error) {
	return []byte("signed"), nil
}

func (fakeWalletProvider) Address(context.Context) (string, error) {
	return "0x3000000000000000000000000000000000000003", nil
}

type fakeEthCaller struct{}

func (fakeEthCaller) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	return []byte{}, nil
}
