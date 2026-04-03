package finance

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUSDC(t *testing.T) {
	tests := []struct {
		give    string
		want    int64
		wantErr bool
	}{
		{give: "1.00", want: 1_000_000},
		{give: "0.50", want: 500_000},
		{give: "100", want: 100_000_000},
		{give: "0.000001", want: 1},
		{give: "0", want: 0},
		{give: "0.1234567", wantErr: true},
		{give: "abc", wantErr: true},
		{give: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseUSDC(tt.give)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Int64())
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
		{give: 0, want: "0.00"},
		{give: 1, want: "0.000001"},
		{give: 100_000_000, want: "100.00"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatUSDC(big.NewInt(tt.give))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFloatToMicroUSDC(t *testing.T) {
	tests := []struct {
		give string
		giveF float64
		want int64
	}{
		{give: "1.0", giveF: 1.0, want: 1_000_000},
		{give: "0.5", giveF: 0.5, want: 500_000},
		{give: "100.0", giveF: 100.0, want: 100_000_000},
		{give: "0.000001", giveF: 0.000001, want: 1},
		{give: "0", giveF: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := FloatToMicroUSDC(tt.giveF)
			assert.Equal(t, tt.want, got.Int64())
		})
	}
}

func TestFloatToMicroUSDC_PrecisionSafety(t *testing.T) {
	// This test verifies that FloatToMicroUSDC handles values that would
	// produce incorrect results with the naive int64(amount * 1_000_000) pattern.
	amount := 9007199.254740993 // near float64 precision limit
	got := FloatToMicroUSDC(amount)

	// The result should be reasonable (within 1 micro-unit of expected).
	expected := big.NewInt(9_007_199_254_740)
	diff := new(big.Int).Sub(got, expected)
	diff.Abs(diff)
	assert.True(t, diff.Int64() <= 1, "precision drift too large: got %s, expected %s", got, expected)

	// Verify the old pattern would overflow for extreme values.
	huge := float64(math.MaxInt64) / 500_000 // would overflow int64 with *1_000_000
	gotHuge := FloatToMicroUSDC(huge)
	assert.True(t, gotHuge.Sign() > 0, "should handle large values without overflow")
}
