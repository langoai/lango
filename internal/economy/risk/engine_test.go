package risk

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/wallet"
)

func mockReputation(scores map[string]float64) ReputationQuerier {
	return func(_ context.Context, peerDID string) (float64, error) {
		return scores[peerDID], nil
	}
}

func mockReputationErr(err error) ReputationQuerier {
	return func(_ context.Context, _ string) (float64, error) {
		return 0, err
	}
}

func usdc(n int64) *big.Int {
	return big.NewInt(n * 1_000_000) // 6 decimal places
}

func newTestEngine(t *testing.T, trust float64) *Engine {
	t.Helper()
	rep := mockReputation(map[string]float64{"peer1": trust})
	e, err := New(config.RiskConfig{}, rep)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return e
}

func TestEngine_Assess_StrategyMatrix(t *testing.T) {
	tests := []struct {
		give         string
		giveTrust    float64
		giveAmount   *big.Int
		giveVerify   Verifiability
		wantStrategy Strategy
	}{
		// === High trust (>= 0.8) ===
		// Low value: always DirectPay regardless of verifiability
		{
			give:         "high trust, low value, high verify -> direct pay",
			giveTrust:    0.9,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "high trust, low value, medium verify -> direct pay",
			giveTrust:    0.85,
			giveAmount:   usdc(2),
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "high trust, low value, low verify -> direct pay",
			giveTrust:    0.9,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyDirectPay,
		},
		// High value: forced escrow
		{
			give:         "high trust, high value, high verify -> escrow",
			giveTrust:    0.95,
			giveAmount:   usdc(10),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyEscrow,
		},
		{
			give:         "high trust, high value, low verify -> escrow",
			giveTrust:    0.85,
			giveAmount:   usdc(10),
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyEscrow,
		},

		// === Medium trust (0.5 - 0.8) ===
		// Low value: depends on verifiability
		{
			give:         "medium trust, low value, high verify -> direct pay",
			giveTrust:    0.6,
			giveAmount:   usdc(2),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "medium trust, low value, medium verify -> micro payment",
			giveTrust:    0.65,
			giveAmount:   usdc(3),
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyMicroPayment,
		},
		{
			give:         "medium trust, low value, low verify -> escrow",
			giveTrust:    0.55,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyEscrow,
		},
		// High value: forced escrow
		{
			give:         "medium trust, high value, high verify -> escrow",
			giveTrust:    0.7,
			giveAmount:   usdc(10),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyEscrow,
		},
		{
			give:         "medium trust, high value, low verify -> escrow",
			giveTrust:    0.6,
			giveAmount:   usdc(20),
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyEscrow,
		},

		// === Low trust (< 0.5) ===
		// Low value: depends on verifiability
		{
			give:         "low trust, low value, high verify -> micro payment",
			giveTrust:    0.2,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyMicroPayment,
		},
		{
			give:         "low trust, low value, medium verify -> zk first",
			giveTrust:    0.3,
			giveAmount:   usdc(2),
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyZKFirst,
		},
		{
			give:         "low trust, low value, low verify -> zk first",
			giveTrust:    0.1,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyZKFirst,
		},
		// High value: always ZKEscrow
		{
			give:         "low trust, high value, high verify -> zk escrow",
			giveTrust:    0.2,
			giveAmount:   usdc(50),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyZKEscrow,
		},
		{
			give:         "zero trust, high value, low verify -> zk escrow",
			giveTrust:    0.0,
			giveAmount:   usdc(100),
			giveVerify:   VerifiabilityLow,
			wantStrategy: StrategyZKEscrow,
		},

		// === Boundary: trust thresholds ===
		{
			give:         "exactly high trust threshold -> direct pay",
			giveTrust:    0.8,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "exactly medium trust threshold, high verify -> direct pay",
			giveTrust:    0.5,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "exactly medium trust threshold, medium verify -> micro payment",
			giveTrust:    0.5,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityMedium,
			wantStrategy: StrategyMicroPayment,
		},
		{
			give:         "just below medium trust, high verify -> micro payment",
			giveTrust:    0.49,
			giveAmount:   usdc(1),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyMicroPayment,
		},

		// === Boundary: escrow threshold ===
		// amount > threshold = high value; amount <= threshold = low value
		{
			give:         "medium trust, at escrow threshold -> direct pay (not high value)",
			giveTrust:    0.6,
			giveAmount:   usdc(5),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyDirectPay,
		},
		{
			give:         "medium trust, just above escrow threshold -> escrow",
			giveTrust:    0.6,
			giveAmount:   big.NewInt(5_000_001),
			giveVerify:   VerifiabilityHigh,
			wantStrategy: StrategyEscrow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			rep := mockReputation(map[string]float64{"peer1": tt.giveTrust})
			engine, err := New(config.RiskConfig{}, rep)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			assessment, err := engine.Assess(context.Background(), "peer1", tt.giveAmount, tt.giveVerify)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if assessment.Strategy != tt.wantStrategy {
				t.Errorf("strategy: got %q, want %q", assessment.Strategy, tt.wantStrategy)
			}
		})
	}
}

func TestEngine_Assess_RiskScoreRange(t *testing.T) {
	tests := []struct {
		give       string
		giveTrust  float64
		giveAmount *big.Int
		giveVerify Verifiability
		wantLevel  RiskLevel
	}{
		{
			give:       "high trust, low value, high verify -> low risk",
			giveTrust:  0.95,
			giveAmount: usdc(1),
			giveVerify: VerifiabilityHigh,
			wantLevel:  RiskLow,
		},
		{
			give:       "zero trust, high value, low verify -> critical risk",
			giveTrust:  0.0,
			giveAmount: usdc(100),
			giveVerify: VerifiabilityLow,
			wantLevel:  RiskCritical,
		},
		{
			give:       "medium trust, medium value, medium verify -> medium risk",
			giveTrust:  0.5,
			giveAmount: usdc(3),
			giveVerify: VerifiabilityMedium,
			wantLevel:  RiskMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			rep := mockReputation(map[string]float64{"peer1": tt.giveTrust})
			engine, err := New(config.RiskConfig{}, rep)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			assessment, err := engine.Assess(context.Background(), "peer1", tt.giveAmount, tt.giveVerify)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if assessment.RiskScore < 0 || assessment.RiskScore > 1 {
				t.Errorf("risk score out of range: %f", assessment.RiskScore)
			}
			if assessment.RiskLevel != tt.wantLevel {
				t.Errorf("risk level: got %q, want %q (score=%.3f)", assessment.RiskLevel, tt.wantLevel, assessment.RiskScore)
			}
		})
	}
}

func TestEngine_Assess_Fields(t *testing.T) {
	rep := mockReputation(map[string]float64{"did:test:alice": 0.75})
	engine, err := New(config.RiskConfig{}, rep)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	amount := usdc(3)

	assessment, err := engine.Assess(context.Background(), "did:test:alice", amount, VerifiabilityMedium)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if assessment.PeerDID != "did:test:alice" {
		t.Errorf("PeerDID: got %q, want %q", assessment.PeerDID, "did:test:alice")
	}
	if assessment.Amount.Cmp(amount) != 0 {
		t.Errorf("Amount: got %s, want %s", assessment.Amount, amount)
	}
	if assessment.TrustScore != 0.75 {
		t.Errorf("TrustScore: got %f, want %f", assessment.TrustScore, 0.75)
	}
	if assessment.Verifiability != VerifiabilityMedium {
		t.Errorf("Verifiability: got %q, want %q", assessment.Verifiability, VerifiabilityMedium)
	}
	if len(assessment.Factors) != 3 {
		t.Errorf("Factors count: got %d, want 3", len(assessment.Factors))
	}
	if assessment.Explanation == "" {
		t.Error("Explanation should not be empty")
	}
	if assessment.AssessedAt.IsZero() {
		t.Error("AssessedAt should not be zero")
	}
	// Amount should be a defensive copy.
	amount.SetInt64(0)
	if assessment.Amount.Sign() == 0 {
		t.Error("Amount should be a defensive copy")
	}
}

func TestEngine_Assess_FactorWeights(t *testing.T) {
	engine := newTestEngine(t, 0.5)

	assessment, err := engine.Assess(context.Background(), "peer1", usdc(1), VerifiabilityHigh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantFactors := map[string]float64{
		"trust":         0.40,
		"value":         0.35,
		"verifiability": 0.25,
	}

	for _, f := range assessment.Factors {
		wantWeight, ok := wantFactors[f.Name]
		if !ok {
			t.Errorf("unexpected factor %q", f.Name)
			continue
		}
		if f.Weight != wantWeight {
			t.Errorf("factor %q weight: got %f, want %f", f.Name, f.Weight, wantWeight)
		}
		if f.Value < 0 || f.Value > 1 {
			t.Errorf("factor %q value out of range: %f", f.Name, f.Value)
		}
		delete(wantFactors, f.Name)
	}
	for name := range wantFactors {
		t.Errorf("missing factor %q", name)
	}
}

func TestEngine_Assess_ReputationError(t *testing.T) {
	dbErr := errors.New("db connection lost")
	rep := mockReputationErr(dbErr)
	engine, err := New(config.RiskConfig{}, rep)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = engine.Assess(context.Background(), "peer1", usdc(1), VerifiabilityHigh)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Errorf("error should wrap original: got %v", err)
	}
}

func TestEngine_Assess_CustomConfig(t *testing.T) {
	rep := mockReputation(map[string]float64{"peer1": 0.7})
	engine, err := New(config.RiskConfig{
		HighTrustScore:   0.7,
		MediumTrustScore: 0.4,
		EscrowThreshold:  "10.00",
	}, rep)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	assessment, err := engine.Assess(context.Background(), "peer1", usdc(5), VerifiabilityHigh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With custom config, 0.7 meets high trust threshold and 5 USDC <= 10 USDC threshold -> DirectPay
	if assessment.Strategy != StrategyDirectPay {
		t.Errorf("strategy: got %q, want %q", assessment.Strategy, StrategyDirectPay)
	}
}

func TestClassifyRisk(t *testing.T) {
	tests := []struct {
		give      float64
		wantLevel RiskLevel
	}{
		{give: 0.0, wantLevel: RiskLow},
		{give: 0.10, wantLevel: RiskLow},
		{give: 0.29, wantLevel: RiskLow},
		{give: 0.30, wantLevel: RiskMedium},
		{give: 0.59, wantLevel: RiskMedium},
		{give: 0.60, wantLevel: RiskHigh},
		{give: 0.84, wantLevel: RiskHigh},
		{give: 0.85, wantLevel: RiskCritical},
		{give: 1.0, wantLevel: RiskCritical},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score_%.2f", tt.give), func(t *testing.T) {
			got := classifyRisk(tt.give)
			if got != tt.wantLevel {
				t.Errorf("classifyRisk(%.2f): got %q, want %q", tt.give, got, tt.wantLevel)
			}
		})
	}
}

func TestTrustFactor(t *testing.T) {
	tests := []struct {
		give     float64
		wantRisk float64
	}{
		{give: 1.0, wantRisk: 0.0},
		{give: 0.8, wantRisk: 0.2},
		{give: 0.5, wantRisk: 0.5},
		{give: 0.0, wantRisk: 1.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("trust_%.1f", tt.give), func(t *testing.T) {
			f := trustFactor(tt.give)
			if f.Name != "trust" {
				t.Errorf("name: got %q, want %q", f.Name, "trust")
			}
			if f.Weight != 0.4 {
				t.Errorf("weight: got %f, want 0.4", f.Weight)
			}
			if fmt.Sprintf("%.2f", f.Value) != fmt.Sprintf("%.2f", tt.wantRisk) {
				t.Errorf("value: got %f, want %f", f.Value, tt.wantRisk)
			}
		})
	}
}

func TestVerifiabilityFactor(t *testing.T) {
	tests := []struct {
		give     Verifiability
		wantRisk float64
	}{
		{give: VerifiabilityHigh, wantRisk: 0.1},
		{give: VerifiabilityMedium, wantRisk: 0.5},
		{give: VerifiabilityLow, wantRisk: 0.9},
		{give: Verifiability("unknown"), wantRisk: 0.9},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			f := verifiabilityFactor(tt.give)
			if f.Name != "verifiability" {
				t.Errorf("name: got %q, want %q", f.Name, "verifiability")
			}
			if f.Weight != 0.25 {
				t.Errorf("weight: got %f, want 0.25", f.Weight)
			}
			if f.Value != tt.wantRisk {
				t.Errorf("value: got %f, want %f", f.Value, tt.wantRisk)
			}
		})
	}
}

func TestAmountFactor(t *testing.T) {
	threshold := usdc(5)

	tests := []struct {
		give    *big.Int
		wantMin float64
		wantMax float64
	}{
		{give: big.NewInt(0), wantMin: 0.0, wantMax: 0.0},
		{give: nil, wantMin: 0.0, wantMax: 0.0},
		{give: big.NewInt(-1), wantMin: 0.0, wantMax: 0.0},
		{give: usdc(1), wantMin: 0.1, wantMax: 0.3},
		{give: usdc(5), wantMin: 0.45, wantMax: 0.55},
		{give: usdc(10), wantMin: 0.6, wantMax: 0.7},
		{give: usdc(100), wantMin: 0.9, wantMax: 1.0},
	}

	for _, tt := range tests {
		label := "nil"
		if tt.give != nil {
			label = tt.give.String()
		}
		t.Run(label, func(t *testing.T) {
			f := amountFactor(tt.give, threshold)
			if f.Value < tt.wantMin || f.Value > tt.wantMax {
				t.Errorf("amountFactor(%s): got %f, want [%f, %f]", label, f.Value, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		give float64
		want float64
	}{
		{give: -1.0, want: 0.0},
		{give: 0.0, want: 0.0},
		{give: 0.5, want: 0.5},
		{give: 1.0, want: 1.0},
		{give: 2.0, want: 1.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.1f", tt.give), func(t *testing.T) {
			got := clamp(tt.give)
			if got != tt.want {
				t.Errorf("clamp(%.1f): got %f, want %f", tt.give, got, tt.want)
			}
		})
	}
}

func TestParseUSDC(t *testing.T) {
	tests := []struct {
		give    string
		want    *big.Int
		wantErr bool
	}{
		{give: "", wantErr: true},
		{give: "not-a-number", wantErr: true},
		{give: "5.00", want: usdc(5)},
		{give: "10.50", want: big.NewInt(10_500_000)},
		{give: "0.01", want: big.NewInt(10_000)},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := wallet.ParseUSDC(tt.give)
			if tt.wantErr {
				if err == nil {
					t.Errorf("wallet.ParseUSDC(%q): want error, got %s", tt.give, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("wallet.ParseUSDC(%q): unexpected error: %v", tt.give, err)
			}
			if got.Cmp(tt.want) != 0 {
				t.Errorf("wallet.ParseUSDC(%q): got %s, want %s", tt.give, got, tt.want)
			}
		})
	}
}
