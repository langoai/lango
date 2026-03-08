package risk

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustFactor_Inversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      float64
		wantValue float64
	}{
		{give: 0.0, wantValue: 1.0},
		{give: 0.2, wantValue: 0.8},
		{give: 0.5, wantValue: 0.5},
		{give: 0.8, wantValue: 0.2},
		{give: 1.0, wantValue: 0.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("trust_%.1f", tt.give), func(t *testing.T) {
			t.Parallel()
			f := trustFactor(tt.give)
			assert.Equal(t, "trust", f.Name)
			assert.InDelta(t, 0.4, f.Weight, 0.001)
			assert.InDelta(t, tt.wantValue, f.Value, 0.001)
		})
	}
}

func TestTrustFactor_ClampsBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      float64
		wantValue float64
	}{
		// trust > 1.0 yields negative before clamp -> clamped to 0.0
		{give: 1.5, wantValue: 0.0},
		// trust < 0.0 yields > 1.0 before clamp -> clamped to 1.0
		{give: -0.5, wantValue: 1.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("trust_%.1f", tt.give), func(t *testing.T) {
			t.Parallel()
			f := trustFactor(tt.give)
			assert.InDelta(t, tt.wantValue, f.Value, 0.001)
		})
	}
}

func TestAmountFactor_Sigmoid(t *testing.T) {
	t.Parallel()

	threshold := big.NewInt(5_000_000) // 5 USDC

	tests := []struct {
		give     string
		giveAmt  *big.Int
		wantMin  float64
		wantMax  float64
		wantName string
	}{
		// nil amount -> 0
		{give: "nil", giveAmt: nil, wantMin: 0.0, wantMax: 0.001, wantName: "value"},
		// zero amount -> 0
		{give: "zero", giveAmt: big.NewInt(0), wantMin: 0.0, wantMax: 0.001, wantName: "value"},
		// negative amount -> 0
		{give: "negative", giveAmt: big.NewInt(-100), wantMin: 0.0, wantMax: 0.001, wantName: "value"},
		// ratio = 1/5 = 0.2, sigmoid = 0.2/1.2 = 0.1667
		{give: "1USDC", giveAmt: big.NewInt(1_000_000), wantMin: 0.15, wantMax: 0.18, wantName: "value"},
		// ratio = 5/5 = 1.0, sigmoid = 1/2 = 0.5
		{give: "equal_threshold", giveAmt: big.NewInt(5_000_000), wantMin: 0.49, wantMax: 0.51, wantName: "value"},
		// ratio = 10/5 = 2.0, sigmoid = 2/3 = 0.6667
		{give: "double_threshold", giveAmt: big.NewInt(10_000_000), wantMin: 0.65, wantMax: 0.68, wantName: "value"},
		// ratio = 100/5 = 20.0, sigmoid = 20/21 ≈ 0.952
		{give: "20x_threshold", giveAmt: big.NewInt(100_000_000), wantMin: 0.95, wantMax: 0.96, wantName: "value"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			f := amountFactor(tt.giveAmt, threshold)
			assert.Equal(t, tt.wantName, f.Name)
			assert.InDelta(t, 0.35, f.Weight, 0.001)
			assert.GreaterOrEqual(t, f.Value, tt.wantMin, "value too low for %s", tt.give)
			assert.LessOrEqual(t, f.Value, tt.wantMax, "value too high for %s", tt.give)
		})
	}
}

func TestAmountFactor_NilThreshold(t *testing.T) {
	t.Parallel()

	// nil threshold produces zero value regardless of amount.
	f := amountFactor(big.NewInt(1_000_000), nil)
	assert.InDelta(t, 0.0, f.Value, 0.001)
}

func TestAmountFactor_ZeroThreshold(t *testing.T) {
	t.Parallel()

	// zero threshold produces zero value to avoid division by zero.
	f := amountFactor(big.NewInt(1_000_000), big.NewInt(0))
	assert.InDelta(t, 0.0, f.Value, 0.001)
}

func TestAmountFactor_SigmoidMonotonicity(t *testing.T) {
	t.Parallel()

	// Verify the sigmoid curve is monotonically increasing.
	threshold := big.NewInt(5_000_000)
	amounts := []int64{100_000, 500_000, 1_000_000, 5_000_000, 10_000_000, 50_000_000}

	var prev float64
	for _, amt := range amounts {
		f := amountFactor(big.NewInt(amt), threshold)
		assert.GreaterOrEqual(t, f.Value, prev, "sigmoid must be monotonically increasing at amount %d", amt)
		prev = f.Value
	}
}

func TestVerifiabilityFactor_AllLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      Verifiability
		wantValue float64
	}{
		{give: VerifiabilityHigh, wantValue: 0.1},
		{give: VerifiabilityMedium, wantValue: 0.5},
		{give: VerifiabilityLow, wantValue: 0.9},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			t.Parallel()
			f := verifiabilityFactor(tt.give)
			assert.Equal(t, "verifiability", f.Name)
			assert.InDelta(t, 0.25, f.Weight, 0.001)
			assert.InDelta(t, tt.wantValue, f.Value, 0.001)
		})
	}
}

func TestVerifiabilityFactor_UnknownDefaultsToLow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give Verifiability
	}{
		{give: Verifiability("unknown")},
		{give: Verifiability("")},
		{give: Verifiability("something_else")},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			t.Parallel()
			f := verifiabilityFactor(tt.give)
			// Unknown verifiability defaults to highest risk (0.9).
			assert.InDelta(t, 0.9, f.Value, 0.001)
		})
	}
}

func TestComputeRiskScore_WeightedAverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		factors   []Factor
		wantScore float64
	}{
		{
			give:      "single factor",
			factors:   []Factor{{Name: "a", Value: 0.6, Weight: 1.0}},
			wantScore: 0.6,
		},
		{
			give: "equal weights",
			factors: []Factor{
				{Name: "a", Value: 0.2, Weight: 1.0},
				{Name: "b", Value: 0.8, Weight: 1.0},
			},
			// (0.2*1.0 + 0.8*1.0) / (1.0 + 1.0) = 0.5
			wantScore: 0.5,
		},
		{
			give: "real factor weights",
			factors: []Factor{
				{Name: "trust", Value: 0.5, Weight: 0.4},
				{Name: "value", Value: 0.5, Weight: 0.35},
				{Name: "verifiability", Value: 0.5, Weight: 0.25},
			},
			// all values 0.5, any weight -> 0.5
			wantScore: 0.5,
		},
		{
			give: "unequal weights",
			factors: []Factor{
				{Name: "a", Value: 1.0, Weight: 0.4},
				{Name: "b", Value: 0.0, Weight: 0.35},
				{Name: "c", Value: 0.0, Weight: 0.25},
			},
			// (1.0*0.4 + 0.0*0.35 + 0.0*0.25) / (0.4+0.35+0.25) = 0.4/1.0 = 0.4
			wantScore: 0.4,
		},
		{
			give:      "empty factors",
			factors:   []Factor{},
			wantScore: 0.0,
		},
		{
			give:      "nil factors",
			factors:   nil,
			wantScore: 0.0,
		},
		{
			give: "all zero weights",
			factors: []Factor{
				{Name: "a", Value: 0.5, Weight: 0.0},
				{Name: "b", Value: 0.8, Weight: 0.0},
			},
			wantScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			score := computeRiskScore(tt.factors)
			assert.InDelta(t, tt.wantScore, score, 0.001)
		})
	}
}

func TestComputeRiskScore_OutputClamped(t *testing.T) {
	t.Parallel()

	// Even with extreme values, output is in [0, 1].
	tests := []struct {
		give    string
		factors []Factor
	}{
		{
			give:    "all max",
			factors: []Factor{{Name: "a", Value: 1.0, Weight: 1.0}},
		},
		{
			give:    "all min",
			factors: []Factor{{Name: "a", Value: 0.0, Weight: 1.0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			score := computeRiskScore(tt.factors)
			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestClassifyRisk_Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      float64
		wantLevel RiskLevel
	}{
		// Low: [0, 0.3)
		{give: 0.0, wantLevel: RiskLow},
		{give: 0.15, wantLevel: RiskLow},
		{give: 0.29, wantLevel: RiskLow},
		{give: 0.299, wantLevel: RiskLow},

		// Medium: [0.3, 0.6)
		{give: 0.3, wantLevel: RiskMedium},
		{give: 0.30, wantLevel: RiskMedium},
		{give: 0.45, wantLevel: RiskMedium},
		{give: 0.59, wantLevel: RiskMedium},
		{give: 0.599, wantLevel: RiskMedium},

		// High: [0.6, 0.85)
		{give: 0.6, wantLevel: RiskHigh},
		{give: 0.60, wantLevel: RiskHigh},
		{give: 0.7, wantLevel: RiskHigh},
		{give: 0.84, wantLevel: RiskHigh},
		{give: 0.849, wantLevel: RiskHigh},

		// Critical: [0.85, 1.0]
		{give: 0.85, wantLevel: RiskCritical},
		{give: 0.9, wantLevel: RiskCritical},
		{give: 1.0, wantLevel: RiskCritical},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score_%.3f", tt.give), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantLevel, classifyRisk(tt.give))
		})
	}
}

func TestComputeFactors_ProducesThreeFactors(t *testing.T) {
	t.Parallel()

	threshold := big.NewInt(5_000_000)
	amount := big.NewInt(1_000_000)

	factors := computeFactors(0.7, amount, threshold, VerifiabilityMedium)
	require.Len(t, factors, 3)

	names := make(map[string]bool, 3)
	for _, f := range factors {
		names[f.Name] = true
		assert.GreaterOrEqual(t, f.Value, 0.0)
		assert.LessOrEqual(t, f.Value, 1.0)
		assert.Greater(t, f.Weight, 0.0)
	}
	assert.True(t, names["trust"])
	assert.True(t, names["value"])
	assert.True(t, names["verifiability"])
}

func TestComputeFactors_WeightsSumToOne(t *testing.T) {
	t.Parallel()

	factors := computeFactors(0.5, big.NewInt(1_000_000), big.NewInt(5_000_000), VerifiabilityHigh)

	var totalWeight float64
	for _, f := range factors {
		totalWeight += f.Weight
	}
	assert.InDelta(t, 1.0, totalWeight, 0.001)
}

func TestComputeRiskScore_RealFactors(t *testing.T) {
	t.Parallel()

	// High trust (0.9), low amount (1 USDC), high verifiability
	// trust factor value = 1 - 0.9 = 0.1, weight 0.4
	// amount factor value = (1/5) / (1 + 1/5) = 0.2/1.2 ≈ 0.1667, weight 0.35
	// verifiability factor value = 0.1, weight 0.25
	//
	// weighted sum = 0.1*0.4 + 0.1667*0.35 + 0.1*0.25
	//             = 0.04 + 0.05833 + 0.025
	//             = 0.12333
	// total weight = 0.4 + 0.35 + 0.25 = 1.0
	// score = 0.12333

	factors := computeFactors(0.9, big.NewInt(1_000_000), big.NewInt(5_000_000), VerifiabilityHigh)
	score := computeRiskScore(factors)

	assert.InDelta(t, 0.123, score, 0.01)
	assert.Equal(t, RiskLow, classifyRisk(score))
}

func TestComputeRiskScore_WorstCase(t *testing.T) {
	t.Parallel()

	// Zero trust, huge amount, low verifiability
	// trust factor value = 1 - 0 = 1.0, weight 0.4
	// amount factor: ratio = 100/5 = 20, sigmoid = 20/21 ≈ 0.952, weight 0.35
	// verifiability factor value = 0.9, weight 0.25
	//
	// weighted sum = 1.0*0.4 + 0.952*0.35 + 0.9*0.25
	//             = 0.4 + 0.3333 + 0.225
	//             = 0.9583
	// score ≈ 0.958

	factors := computeFactors(0.0, big.NewInt(100_000_000), big.NewInt(5_000_000), VerifiabilityLow)
	score := computeRiskScore(factors)

	assert.InDelta(t, 0.958, score, 0.01)
	assert.Equal(t, RiskCritical, classifyRisk(score))
}

func TestClamp_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give float64
		want float64
	}{
		{give: -100.0, want: 0.0},
		{give: -0.001, want: 0.0},
		{give: 0.0, want: 0.0},
		{give: 0.5, want: 0.5},
		{give: 1.0, want: 1.0},
		{give: 1.001, want: 1.0},
		{give: 100.0, want: 1.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.3f", tt.give), func(t *testing.T) {
			t.Parallel()
			assert.InDelta(t, tt.want, clamp(tt.give), 0.001)
		})
	}
}
