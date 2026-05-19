package app

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/budget"
	sa "github.com/langoai/lango/internal/smartaccount"
)

func TestWave43ToOnChainPolicy_SelectorEdgesAndZeroTimes(t *testing.T) {
	t.Parallel()

	got := toOnChainPolicy(sa.SessionPolicy{
		AllowedFunctions: []string{
			"a9059cbb",
			"0x1234567890",
			"0x123456",
			"0x",
			"zzzzzzzz",
			"0x0",
		},
	})

	assert.Equal(t, big.NewInt(0), wave43PolicyField(t, got, "SpendLimit"))
	assert.Equal(t, big.NewInt(0), wave43PolicyField(t, got, "SpentAmount"))
	assert.Equal(t, big.NewInt(time.Time{}.Unix()), wave43PolicyField(t, got, "ValidAfter"))
	assert.Equal(t, big.NewInt(time.Time{}.Unix()), wave43PolicyField(t, got, "ValidUntil"))
	assert.Equal(t, true, wave43PolicyField(t, got, "Active"))

	selectors := wave43PolicyField(t, got, "AllowedFunctions").([][4]byte)
	require.Len(t, selectors, 2)
	assert.Equal(t, [4]byte{0xa9, 0x05, 0x9c, 0xbb}, selectors[0])
	assert.Equal(t, [4]byte{0x12, 0x34, 0x56, 0x78}, selectors[1])
}

func TestWave43InitSmartAccount_LightweightWiringSkipsMissingOptionalDependencies(t *testing.T) {
	cfg := wave43SmartAccountConfig()
	cfg.SmartAccount.Modules.SpendingHookAddress = "0x3000000000000000000000000000000000000003"
	cfg.SmartAccount.Modules.EscrowExecutorAddress = "0x4000000000000000000000000000000000000004"
	cfg.SmartAccount.Paymaster.Enabled = true
	cfg.SmartAccount.Paymaster.Mode = "permit"
	cfg.SmartAccount.Paymaster.Provider = "circle"
	cfg.SmartAccount.Paymaster.PaymasterAddress = "0x5000000000000000000000000000000000000005"
	cfg.SmartAccount.Paymaster.TokenAddress = "0x6000000000000000000000000000000000000006"

	result := initSmartAccount(cfg, &paymentComponents{chainID: 84532}, nil, nil)

	require.NotNil(t, result)
	require.NotNil(t, result.components)
	assert.Empty(t, result.lifecycle, "no sentinel guard lifecycle entry should be created without sentinel and bus")
	assert.NotNil(t, result.components.manager)
	assert.NotNil(t, result.components.SessionManager())
	assert.NotNil(t, result.components.PolicyEngine())
	assert.NotNil(t, result.components.BundlerClient())
	assert.NotNil(t, result.components.OnChainTracker())
	assert.Nil(t, result.components.PaymasterProvider(), "permit paymaster requires wallet and RPC dependencies")
	assert.Nil(t, result.components.sessionGuard)

	modules := result.components.ModuleRegistry().List()
	require.Len(t, modules, 2)
	spendingHook := common.HexToAddress(cfg.SmartAccount.Modules.SpendingHookAddress)
	escrowExecutor := common.HexToAddress(cfg.SmartAccount.Modules.EscrowExecutorAddress)
	assertRegisteredModule(t, result.components.ModuleRegistry(), spendingHook, "LangoSpendingHook", sa.ModuleTypeHook)
	assertRegisteredModule(t, result.components.ModuleRegistry(), escrowExecutor, "LangoEscrowExecutor", sa.ModuleTypeExecutor)
}

func TestWave43InitSmartAccount_OnChainTrackerRecordsWithoutBudgetEngine(t *testing.T) {
	cfg := wave43SmartAccountConfig()

	result := initSmartAccount(cfg, &paymentComponents{chainID: 84532}, &economyComponents{}, nil)

	require.NotNil(t, result)
	tracker := result.components.OnChainTracker()
	require.NotNil(t, tracker)

	tracker.Record("session-without-budget", big.NewInt(7))

	assert.Equal(t, big.NewInt(7), tracker.GetSpent("session-without-budget"))
	assert.Empty(t, result.lifecycle)
	assert.Nil(t, result.components.sessionGuard)
}

func TestWave43InitSmartAccount_OnChainTrackerSyncsToBudgetEngineWhenPresent(t *testing.T) {
	cfg := wave43SmartAccountConfig()
	store := budget.NewStore()
	engine, err := budget.NewEngine(store, config.BudgetConfig{})
	require.NoError(t, err)
	_, err = engine.Allocate("session-with-budget", big.NewInt(100))
	require.NoError(t, err)

	result := initSmartAccount(cfg, &paymentComponents{chainID: 84532}, &economyComponents{budgetEngine: engine}, nil)

	require.NotNil(t, result)
	tracker := result.components.OnChainTracker()
	require.NotNil(t, tracker)

	tracker.Record("session-with-budget", big.NewInt(12))

	assert.Equal(t, big.NewInt(12), tracker.GetSpent("session-with-budget"))
	synced, err := store.Get("session-with-budget")
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(12), synced.Spent)
	require.Len(t, synced.Entries, 1)
	assert.Equal(t, "on-chain spend sync", synced.Entries[0].Reason)
	assert.Equal(t, big.NewInt(12), synced.Entries[0].Amount)
}

func wave43SmartAccountConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.SmartAccount.Enabled = true
	cfg.SmartAccount.EntryPointAddress = "0x1000000000000000000000000000000000000001"
	cfg.SmartAccount.FactoryAddress = "0x2000000000000000000000000000000000000002"
	cfg.SmartAccount.SafeSingletonAddress = ""
	cfg.SmartAccount.Safe7579Address = "0x7000000000000000000000000000000000000007"
	cfg.SmartAccount.FallbackHandler = "0x8000000000000000000000000000000000000008"
	cfg.SmartAccount.BundlerURL = "http://127.0.0.1:4337"
	cfg.SmartAccount.Session.MaxDuration = time.Hour
	cfg.SmartAccount.Session.MaxActiveKeys = 2
	return cfg
}

func wave43PolicyField(t *testing.T, got interface{}, name string) interface{} {
	t.Helper()

	value := reflect.ValueOf(got)
	require.Equal(t, reflect.Struct, value.Kind())
	field := value.FieldByName(name)
	require.Truef(t, field.IsValid(), "missing policy field %q", name)
	return field.Interface()
}
