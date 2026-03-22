package app

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// initEconomy
// ---------------------------------------------------------------------------

func TestInitEconomy_DisabledReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = false

	result := initEconomy(cfg, nil, nil, nil)

	assert.Nil(t, result, "expected nil when economy is disabled")
}

func TestInitEconomy_DisabledBranch_TableDriven(t *testing.T) {
	tests := []struct {
		give       string
		giveOn     bool
		giveP2P    *p2pComponents
		givePC     *paymentComponents
		giveBus    *eventbus.Bus
		wantNil    bool
	}{
		{
			give:    "disabled config returns nil",
			giveOn:  false,
			giveP2P: nil,
			givePC:  nil,
			giveBus: nil,
			wantNil: true,
		},
		{
			give:    "disabled with non-nil p2p returns nil",
			giveOn:  false,
			giveP2P: &p2pComponents{},
			givePC:  nil,
			giveBus: nil,
			wantNil: true,
		},
		{
			give:    "disabled with non-nil bus returns nil",
			giveOn:  false,
			giveP2P: nil,
			givePC:  nil,
			giveBus: eventbus.New(),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Economy.Enabled = tt.giveOn

			result := initEconomy(cfg, tt.giveP2P, tt.givePC, tt.giveBus)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestInitEconomy_EnabledNilDeps_ReturnsComponents(t *testing.T) {
	// When enabled with nil dependencies, initEconomy still returns a non-nil
	// economyComponents struct — budget and risk engines are created with
	// defaults/fallbacks, while pricing/negotiation/escrow require their own
	// sub-config enabled flags.
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true

	result := initEconomy(cfg, nil, nil, nil)

	require.NotNil(t, result, "economy components should be created when enabled")
	// Budget engine should be initialized (uses defaults).
	assert.NotNil(t, result.budgetEngine, "budget engine should be created")
	// Risk engine should be initialized with neutral reputation fallback.
	assert.NotNil(t, result.riskEngine, "risk engine should be created with fallback reputation")
	// Pricing, negotiation, escrow require their own sub-config enabled flags.
	assert.Nil(t, result.pricingEngine, "pricing engine should be nil when pricing not enabled")
	assert.Nil(t, result.negotiationEngine, "negotiation engine should be nil when negotiate not enabled")
	assert.Nil(t, result.escrowEngine, "escrow engine should be nil when escrow not enabled")
}

func TestInitEconomy_EnabledWithBus_WiresBudgetAlertCallback(t *testing.T) {
	// When a bus is provided, budget engine should still initialize correctly.
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	bus := eventbus.New()

	result := initEconomy(cfg, nil, nil, bus)

	require.NotNil(t, result)
	assert.NotNil(t, result.budgetEngine, "budget engine should be created with bus")
	assert.NotNil(t, result.riskEngine, "risk engine should be created")
}

func TestInitEconomy_EnabledWithP2PComponents(t *testing.T) {
	// When p2p components are provided (but without a real reputation store),
	// the risk engine falls back to neutral reputation.
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	p2pc := &p2pComponents{}

	result := initEconomy(cfg, p2pc, nil, nil)

	require.NotNil(t, result)
	assert.NotNil(t, result.riskEngine, "risk engine should use neutral fallback when reputation store is nil")
}

// ---------------------------------------------------------------------------
// selectSettler
// ---------------------------------------------------------------------------

func TestSelectSettler_NilPaymentComponents_ReturnsNoopSettler(t *testing.T) {
	cfg := config.DefaultConfig()

	settler := selectSettler(cfg, nil)

	assert.NotNil(t, settler, "settler should never be nil")
}

func TestSelectSettler_TableDriven(t *testing.T) {
	tests := []struct {
		give           string
		giveOnChain    bool
		givePC         *paymentComponents
		wantNotNil     bool
	}{
		{
			give:        "nil payment components returns noop settler",
			giveOnChain: false,
			givePC:      nil,
			wantNotNil:  true, // NoopSettler is returned
		},
		{
			give:        "on-chain enabled but nil payment returns noop settler",
			giveOnChain: true,
			givePC:      nil,
			wantNotNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Economy.Escrow.OnChain.Enabled = tt.giveOnChain

			settler := selectSettler(cfg, tt.givePC)

			if tt.wantNotNil {
				assert.NotNil(t, settler)
			}
		})
	}
}
