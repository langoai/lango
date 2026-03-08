package risk

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/config"
)

func newStrategyEngine(t *testing.T) *Engine {
	t.Helper()
	// Default thresholds: highTrust=0.8, medTrust=0.5, escrowThreshold=5 USDC
	e, err := New(config.RiskConfig{}, func(_ context.Context, _ string) (float64, error) {
		return 0, nil
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return e
}

// lowAmount is below the default escrow threshold (5 USDC).
func lowAmount() *big.Int { return big.NewInt(1_000_000) } // 1 USDC

// highAmount is above the default escrow threshold (5 USDC).
func highAmount() *big.Int { return big.NewInt(10_000_000) } // 10 USDC

func TestSelectStrategy_LowValueMatrix(t *testing.T) {
	t.Parallel()

	engine := newStrategyEngine(t)
	amount := lowAmount()

	// 3 trust levels x 3 verifiability levels = 9 combinations
	tests := []struct {
		give         string
		giveTrust    float64
		giveVerify   Verifiability
		wantStrategy Strategy
	}{
		// === High trust (>= 0.8): always DirectPay for low value ===
		{
			give:         "high trust + high verify",
			giveTrust:    0.9,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "high trust + medium verify",
			giveTrust:    0.85,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "high trust + low verify",
			giveTrust:    0.95,
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyDirectPay,
		},

		// === Medium trust (>= 0.5, < 0.8): depends on verifiability ===
		{
			give:         "medium trust + high verify",
			giveTrust:    0.6,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "medium trust + medium verify",
			giveTrust:    0.65,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyMicroPayment,
		},
		{
			give:         "medium trust + low verify",
			giveTrust:    0.55,
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyEscrow,
		},

		// === Low trust (< 0.5): depends on verifiability ===
		{
			give:         "low trust + high verify",
			giveTrust:    0.3,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyMicroPayment,
		},
		{
			give:         "low trust + medium verify",
			giveTrust:    0.2,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyZKFirst,
		},
		{
			give:         "low trust + low verify",
			giveTrust:    0.1,
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyZKFirst,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := engine.selectStrategy(tt.giveTrust, amount, tt.giveVerify)
			assert.Equal(t, tt.wantStrategy, got)
		})
	}
}

func TestSelectStrategy_HighValueMatrix(t *testing.T) {
	t.Parallel()

	engine := newStrategyEngine(t)
	amount := highAmount()

	tests := []struct {
		give         string
		giveTrust    float64
		giveVerify   Verifiability
		wantStrategy Strategy
	}{
		// High-value forces escrow-based strategies regardless of verifiability.

		// === High trust (>= 0.8): Escrow ===
		{
			give:         "high trust + high verify",
			giveTrust:    0.9,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyEscrow,
		},
		{
			give:         "high trust + medium verify",
			giveTrust:    0.85,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyEscrow,
		},
		{
			give:         "high trust + low verify",
			giveTrust:    0.95,
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyEscrow,
		},

		// === Medium trust (>= 0.5): Escrow ===
		{
			give:         "medium trust + high verify",
			giveTrust:    0.7,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyEscrow,
		},
		{
			give:         "medium trust + medium verify",
			giveTrust:    0.6,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyEscrow,
		},
		{
			give:         "medium trust + low verify",
			giveTrust:    0.55,
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyEscrow,
		},

		// === Low trust (< 0.5): ZKEscrow ===
		{
			give:         "low trust + high verify",
			giveTrust:    0.3,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyZKEscrow,
		},
		{
			give:         "low trust + medium verify",
			giveTrust:    0.2,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyZKEscrow,
		},
		{
			give:         "low trust + low verify",
			giveTrust:    0.0,
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyZKEscrow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := engine.selectStrategy(tt.giveTrust, amount, tt.giveVerify)
			assert.Equal(t, tt.wantStrategy, got)
		})
	}
}

func TestSelectStrategy_TrustBoundaries(t *testing.T) {
	t.Parallel()

	engine := newStrategyEngine(t)
	amount := lowAmount()

	tests := []struct {
		give         string
		giveTrust    float64
		giveVerify   Verifiability
		wantStrategy Strategy
	}{
		// Exact high trust boundary (0.8)
		{
			give:         "exactly high trust threshold",
			giveTrust:    0.8,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		// Just below high trust
		{
			give:         "just below high trust",
			giveTrust:    0.79,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay, // medium trust + high verify -> direct pay
		},
		// Exact medium trust boundary (0.5)
		{
			give:         "exactly medium trust threshold + high verify",
			giveTrust:    0.5,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "exactly medium trust threshold + medium verify",
			giveTrust:    0.5,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyMicroPayment,
		},
		{
			give:         "exactly medium trust threshold + low verify",
			giveTrust:    0.5,
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyEscrow,
		},
		// Just below medium trust
		{
			give:         "just below medium trust + high verify",
			giveTrust:    0.49,
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyMicroPayment,
		},
		{
			give:         "just below medium trust + medium verify",
			giveTrust:    0.49,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyZKFirst,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := engine.selectStrategy(tt.giveTrust, amount, tt.giveVerify)
			assert.Equal(t, tt.wantStrategy, got)
		})
	}
}

func TestSelectStrategy_EscrowThresholdBoundary(t *testing.T) {
	t.Parallel()

	engine := newStrategyEngine(t)

	tests := []struct {
		give         string
		giveTrust    float64
		giveAmount   *big.Int
		giveVerify   Verifiability
		wantStrategy Strategy
	}{
		{
			give:         "at escrow threshold (not high value)",
			giveTrust:    0.6,
			giveAmount:   big.NewInt(5_000_000), // exactly 5 USDC = threshold
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay, // not > threshold, so low-value path
		},
		{
			give:         "one above escrow threshold (high value)",
			giveTrust:    0.6,
			giveAmount:   big.NewInt(5_000_001),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyEscrow, // > threshold, medium trust -> escrow
		},
		{
			give:         "one above threshold + low trust",
			giveTrust:    0.3,
			giveAmount:   big.NewInt(5_000_001),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyZKEscrow, // > threshold, low trust -> zk_escrow
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := engine.selectStrategy(tt.giveTrust, tt.giveAmount, tt.giveVerify)
			assert.Equal(t, tt.wantStrategy, got)
		})
	}
}

func TestSelectStrategy_ExtremeTrustValues(t *testing.T) {
	t.Parallel()

	engine := newStrategyEngine(t)
	amount := lowAmount()

	tests := []struct {
		give         string
		giveTrust    float64
		wantStrategy Strategy
	}{
		{
			give:         "perfect trust",
			giveTrust:    1.0,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "zero trust",
			giveTrust:    0.0,
			wantStrategy: StrategyZKFirst, // low trust + low verify default
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := engine.selectStrategy(tt.giveTrust, amount, VerifiabilityLow)
			assert.Equal(t, tt.wantStrategy, got)
		})
	}
}

func TestSelectStrategy_CustomThresholds(t *testing.T) {
	t.Parallel()

	// Custom config with different trust thresholds.
	e, err := New(config.RiskConfig{
		HighTrustScore:   0.9,
		MediumTrustScore: 0.6,
		EscrowThreshold:  "20.00",
	}, func(_ context.Context, _ string) (float64, error) {
		return 0, nil
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	amount := big.NewInt(1_000_000) // 1 USDC, well below 20 USDC threshold

	tests := []struct {
		give         string
		giveTrust    float64
		giveVerify   Verifiability
		wantStrategy Strategy
	}{
		{
			give:         "0.85 is medium (not high) with custom thresholds",
			giveTrust:    0.85,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyMicroPayment, // medium trust + medium verify
		},
		{
			give:         "0.9 meets custom high trust",
			giveTrust:    0.9,
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "0.55 is low trust with custom thresholds",
			giveTrust:    0.55,
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyZKFirst, // low trust + medium verify
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := e.selectStrategy(tt.giveTrust, amount, tt.giveVerify)
			assert.Equal(t, tt.wantStrategy, got)
		})
	}
}

func TestSelectStrategy_HighValueIgnoresVerifiability(t *testing.T) {
	t.Parallel()

	engine := newStrategyEngine(t)
	amount := highAmount()

	// For high-value + high trust, verifiability shouldn't matter — always Escrow.
	verifiabilities := []Verifiability{VerifiabilityHigh, VerifiabilityMedium, VerifiabilityLow}
	for _, v := range verifiabilities {
		t.Run("high_trust_"+string(v), func(t *testing.T) {
			t.Parallel()
			got := engine.selectStrategy(0.9, amount, v)
			assert.Equal(t, StrategyEscrow, got)
		})
	}

	// For high-value + low trust, verifiability shouldn't matter — always ZKEscrow.
	for _, v := range verifiabilities {
		t.Run("low_trust_"+string(v), func(t *testing.T) {
			t.Parallel()
			got := engine.selectStrategy(0.2, amount, v)
			assert.Equal(t, StrategyZKEscrow, got)
		})
	}
}
