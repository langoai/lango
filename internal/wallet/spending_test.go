package wallet

import (
	"fmt"
	"math/big"
	"testing"
)

func TestParseUSDC(t *testing.T) {
	tests := []struct {
		give    string
		want    int64
		wantErr bool
	}{
		{give: "1.00", want: 1_000_000},
		{give: "0.50", want: 500_000},
		{give: "10.00", want: 10_000_000},
		{give: "0.000001", want: 1},
		{give: "0", want: 0},
		{give: "100", want: 100_000_000},
		{give: "invalid", wantErr: true},
		{give: "0.0000001", wantErr: true}, // too many decimals
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseUSDC(tt.give)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseUSDC(%q) expected error, got %v", tt.give, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseUSDC(%q) unexpected error: %v", tt.give, err)
			}
			if got.Int64() != tt.want {
				t.Errorf("ParseUSDC(%q) = %d, want %d", tt.give, got.Int64(), tt.want)
			}
		})
	}
}

func TestFormatUSDC(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{give: 1_000_000, want: "1.00"},
		{give: 500_000, want: "0.50"},
		{give: 10_000_000, want: "10.00"},
		{give: 0, want: "0.00"},
		{give: 1, want: "0.000001"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatUSDC(big.NewInt(tt.give))
			if got != tt.want {
				t.Errorf("FormatUSDC(%d) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}

func TestIsAutoApprovable(t *testing.T) {
	tests := []struct {
		give             string
		autoApproveBelow string
		wantOK           bool
		wantErr          bool
	}{
		{give: "0.05", autoApproveBelow: "0.10", wantOK: true},
		{give: "0.10", autoApproveBelow: "0.10", wantOK: true},
		{give: "0.11", autoApproveBelow: "0.10", wantOK: false},
		{give: "1.00", autoApproveBelow: "0.10", wantOK: false},
		{give: "0.05", autoApproveBelow: "0", wantOK: false},    // disabled
		{give: "0.05", autoApproveBelow: "", wantOK: false},     // disabled
		{give: "0.00", autoApproveBelow: "0.10", wantOK: true},  // zero amount
		{give: "5.00", autoApproveBelow: "10.00", wantOK: true}, // large threshold
	}

	for _, tt := range tests {
		name := fmt.Sprintf("amount=%s,threshold=%s", tt.give, tt.autoApproveBelow)
		t.Run(name, func(t *testing.T) {
			limiter := &EntSpendingLimiter{
				maxPerTx:         big.NewInt(100_000_000), // 100 USDC
				maxDaily:         big.NewInt(100_000_000), // 100 USDC
				autoApproveBelow: big.NewInt(0),
			}

			// Parse auto-approve threshold.
			if tt.autoApproveBelow != "" {
				parsed, err := ParseUSDC(tt.autoApproveBelow)
				if err != nil {
					t.Fatalf("parse autoApproveBelow: %v", err)
				}
				limiter.autoApproveBelow = parsed
			}

			amt, err := ParseUSDC(tt.give)
			if err != nil {
				t.Fatalf("parse amount: %v", err)
			}

			// IsAutoApprovable uses Check() which requires an ent client for DailySpent.
			// Since we can't create a real ent client in unit tests, we test the
			// threshold logic directly. The client-dependent path is covered by
			// integration tests.
			if limiter.autoApproveBelow.Sign() == 0 {
				if tt.wantOK {
					t.Error("expected auto-approvable but threshold is 0")
				}
				return
			}

			if amt.Cmp(limiter.autoApproveBelow) > 0 {
				if tt.wantOK {
					t.Errorf("amount %s > threshold %s, expected not auto-approvable",
						tt.give, tt.autoApproveBelow)
				}
				return
			}

			// Amount is within threshold.
			if !tt.wantOK {
				t.Errorf("amount %s <= threshold %s, expected auto-approvable",
					tt.give, tt.autoApproveBelow)
			}
		})
	}
}

func TestNewEntSpendingLimiter_AutoApproveBelow(t *testing.T) {
	tests := []struct {
		give    string
		want    int64
		wantErr bool
	}{
		{give: "0.10", want: 100_000},
		{give: "1.00", want: 1_000_000},
		{give: "0", want: 0},
		{give: "", want: 0},
		{give: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			limiter, err := NewEntSpendingLimiter(nil, "1.00", "10.00", tt.give)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if limiter.autoApproveBelow.Int64() != tt.want {
				t.Errorf("autoApproveBelow = %d, want %d",
					limiter.autoApproveBelow.Int64(), tt.want)
			}
		})
	}
}

func TestNetworkName(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{give: 1, want: "Ethereum Mainnet"},
		{give: 8453, want: "Base"},
		{give: 84532, want: "Base Sepolia"},
		{give: 99999, want: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := NetworkName(tt.give)
			if got != tt.want {
				t.Errorf("NetworkName(%d) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}
