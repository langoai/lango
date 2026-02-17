package wallet

import (
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
		{give: 500_000, want: "0.500000"},
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
