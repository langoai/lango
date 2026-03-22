package app

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// initSmartAccount
// ---------------------------------------------------------------------------

func TestInitSmartAccount_DisabledReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SmartAccount.Enabled = false

	result := initSmartAccount(cfg, nil, nil, nil)

	assert.Nil(t, result, "expected nil when smart account is disabled")
}

func TestInitSmartAccount_NilPaymentComponentsReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SmartAccount.Enabled = true
	// Even with enabled config, nil payment components should cause early return.

	result := initSmartAccount(cfg, nil, nil, nil)

	assert.Nil(t, result, "expected nil when payment components are nil")
}

func TestInitSmartAccount_IncompleteConfigReturnsNil(t *testing.T) {
	// Smart account is enabled and payment components exist, but required
	// config fields (entryPointAddress, factoryAddress, bundlerURL) are missing.
	cfg := config.DefaultConfig()
	cfg.SmartAccount.Enabled = true
	pc := &paymentComponents{}

	result := initSmartAccount(cfg, pc, nil, nil)

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

			result := initSmartAccount(cfg, tt.givePC, tt.giveEconc, tt.giveBus)

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
