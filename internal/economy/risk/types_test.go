package risk

import (
	"testing"
)

func TestRiskLevel_StringValues(t *testing.T) {
	tests := []struct {
		give string
		want RiskLevel
	}{
		{give: "low", want: RiskLow},
		{give: "medium", want: RiskMedium},
		{give: "high", want: RiskHigh},
		{give: "critical", want: RiskCritical},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if string(tt.want) != tt.give {
				t.Errorf("RiskLevel %q: got %q, want %q", tt.give, string(tt.want), tt.give)
			}
		})
	}
}

func TestStrategy_StringValues(t *testing.T) {
	tests := []struct {
		give string
		want Strategy
	}{
		{give: "direct_pay", want: StrategyDirectPay},
		{give: "micro_payment", want: StrategyMicroPayment},
		{give: "escrow", want: StrategyEscrow},
		{give: "zk_first", want: StrategyZKFirst},
		{give: "zk_escrow", want: StrategyZKEscrow},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if string(tt.want) != tt.give {
				t.Errorf("Strategy %q: got %q, want %q", tt.give, string(tt.want), tt.give)
			}
		})
	}
}

func TestVerifiability_StringValues(t *testing.T) {
	tests := []struct {
		give string
		want Verifiability
	}{
		{give: "high", want: VerifiabilityHigh},
		{give: "medium", want: VerifiabilityMedium},
		{give: "low", want: VerifiabilityLow},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if string(tt.want) != tt.give {
				t.Errorf("Verifiability %q: got %q, want %q", tt.give, string(tt.want), tt.give)
			}
		})
	}
}

func TestFactor_ValueWeightRanges(t *testing.T) {
	tests := []struct {
		give      string
		wantValid bool
		factor    Factor
	}{
		{
			give:      "valid factor",
			wantValid: true,
			factor:    Factor{Name: "trust", Value: 0.5, Weight: 0.3},
		},
		{
			give:      "zero values",
			wantValid: true,
			factor:    Factor{Name: "new_peer", Value: 0.0, Weight: 0.0},
		},
		{
			give:      "max values",
			wantValid: true,
			factor:    Factor{Name: "critical", Value: 1.0, Weight: 1.0},
		},
		{
			give:      "value out of range high",
			wantValid: false,
			factor:    Factor{Name: "bad", Value: 1.5, Weight: 0.5},
		},
		{
			give:      "value out of range negative",
			wantValid: false,
			factor:    Factor{Name: "bad", Value: -0.1, Weight: 0.5},
		},
		{
			give:      "weight out of range high",
			wantValid: false,
			factor:    Factor{Name: "bad", Value: 0.5, Weight: 1.5},
		},
		{
			give:      "weight out of range negative",
			wantValid: false,
			factor:    Factor{Name: "bad", Value: 0.5, Weight: -0.1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			valid := tt.factor.Value >= 0.0 && tt.factor.Value <= 1.0 &&
				tt.factor.Weight >= 0.0 && tt.factor.Weight <= 1.0
			if valid != tt.wantValid {
				t.Errorf("Factor %q: valid=%v, wantValid=%v (value=%.1f, weight=%.1f)",
					tt.give, valid, tt.wantValid, tt.factor.Value, tt.factor.Weight)
			}
		})
	}
}
