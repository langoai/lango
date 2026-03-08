package risk

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

// Risk score formula (with VerifiabilityMedium):
//   score = (1-trust)*0.4 + amountFactor*0.35 + 0.5*0.25
//   where amountFactor = ratio/(1+ratio), ratio = amount/threshold
// Boundaries: Low < 0.3, Medium < 0.6, High < 0.85, Critical >= 0.85.

func TestPolicyAdapter_Recommend(t *testing.T) {
	t.Parallel()

	fullBudget := big.NewInt(100_000_000) // 100 USDC
	defaultCfg := config.RiskConfig{}     // default threshold: 5 USDC

	tests := []struct {
		give         string
		giveTrust    float64
		giveCfg      config.RiskConfig
		giveAmount   *big.Int
		wantMaxSpend *big.Int
		wantDuration time.Duration
		wantApproval bool
	}{
		{
			// trust=0.95, 1 USDC, threshold 5 USDC
			// score = 0.05*0.4 + 0.167*0.35 + 0.125 = 0.203 => RiskLow
			give:         "low risk -> full budget, 24h, no approval",
			giveTrust:    0.95,
			giveCfg:      defaultCfg,
			giveAmount:   big.NewInt(1_000_000),
			wantMaxSpend: big.NewInt(100_000_000),
			wantDuration: 24 * time.Hour,
			wantApproval: false,
		},
		{
			// trust=0.65, 1 USDC, threshold 5 USDC
			// score = 0.35*0.4 + 0.167*0.35 + 0.125 = 0.323 => RiskMedium
			give:         "medium risk -> half budget, 6h, no approval",
			giveTrust:    0.65,
			giveCfg:      defaultCfg,
			giveAmount:   big.NewInt(1_000_000),
			wantMaxSpend: big.NewInt(50_000_000),
			wantDuration: 6 * time.Hour,
			wantApproval: false,
		},
		{
			// trust=0.3, 50 USDC, threshold 5 USDC
			// ratio=10, amountVal=0.909
			// score = 0.7*0.4 + 0.909*0.35 + 0.125 = 0.723 => RiskHigh
			give:         "high risk -> 1/10 budget, 1h, requires approval",
			giveTrust:    0.3,
			giveCfg:      defaultCfg,
			giveAmount:   big.NewInt(50_000_000),
			wantMaxSpend: big.NewInt(10_000_000),
			wantDuration: 1 * time.Hour,
			wantApproval: true,
		},
		{
			// trust=0.0, 500 USDC, threshold 5 USDC
			// ratio=100, amountVal=0.99
			// score = 1.0*0.4 + 0.99*0.35 + 0.125 = 0.872 => RiskCritical
			give:         "critical risk -> zero budget, no duration, requires approval",
			giveTrust:    0.0,
			giveCfg:      defaultCfg,
			giveAmount:   big.NewInt(500_000_000),
			wantMaxSpend: new(big.Int),
			wantDuration: 0,
			wantApproval: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			rep := mockReputation(map[string]float64{"peer1": tt.giveTrust})
			engine, err := New(tt.giveCfg, rep)
			require.NoError(t, err)

			adapter := NewPolicyAdapter(engine, fullBudget)
			rec, err := adapter.Recommend(context.Background(), "peer1", tt.giveAmount)
			require.NoError(t, err)

			assert.Equal(t, 0, rec.MaxSpendLimit.Cmp(tt.wantMaxSpend),
				"MaxSpendLimit: got %s, want %s", rec.MaxSpendLimit, tt.wantMaxSpend)
			assert.Equal(t, tt.wantDuration, rec.MaxDuration)
			assert.Equal(t, tt.wantApproval, rec.RequireApproval)
		})
	}
}

func TestPolicyAdapter_Recommend_ReputationError(t *testing.T) {
	t.Parallel()

	rep := mockReputationErr(assert.AnError)
	engine, err := New(config.RiskConfig{}, rep)
	require.NoError(t, err)

	adapter := NewPolicyAdapter(engine, big.NewInt(100_000_000))
	_, err = adapter.Recommend(context.Background(), "peer1", big.NewInt(100))
	require.Error(t, err)
}

func TestPolicyAdapter_AdaptToRiskPolicyFunc(t *testing.T) {
	t.Parallel()

	// AdaptToRiskPolicyFunc passes fullBudget as the amount to Recommend.
	// With trust=0.9, fullBudget=1 USDC (1_000_000), default threshold (5 USDC):
	//   score = 0.1*0.4 + 0.167*0.35 + 0.125 = 0.223 => RiskLow
	rep := mockReputation(map[string]float64{"peer1": 0.9})
	engine, err := New(config.RiskConfig{}, rep)
	require.NoError(t, err)

	fullBudget := big.NewInt(1_000_000) // 1 USDC
	adapter := NewPolicyAdapter(engine, fullBudget)
	fn := adapter.AdaptToRiskPolicyFunc()

	rec, err := fn(context.Background(), "peer1")
	require.NoError(t, err)
	require.NotNil(t, rec)

	// Low risk -> full budget.
	assert.Equal(t, 0, rec.MaxSpendLimit.Cmp(fullBudget))
	assert.Equal(t, 24*time.Hour, rec.MaxDuration)
	assert.False(t, rec.RequireApproval)
}
