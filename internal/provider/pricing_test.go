package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriceFor_KnownPrefix(t *testing.T) {
	tests := []struct {
		give      string
		wantFound bool
	}{
		{"claude-opus-4-6", true},
		{"claude-sonnet-4-6-20251119", true},
		{"claude-haiku-4-5", true},
		{"gemini-2.5-pro", true},
		{"gemini-2.5-flash-001", true},
		{"gpt-4o", true},
		{"gpt-4o-mini", true},
		{"unknown-model-x", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			_, ok := PriceFor(tt.give)
			assert.Equal(t, tt.wantFound, ok)
		})
	}
}

func TestPriceFor_CaseInsensitive(t *testing.T) {
	_, ok := PriceFor("CLAUDE-OPUS-4-6")
	assert.True(t, ok)
}

func TestEstimateCostUSD_Computation(t *testing.T) {
	// claude-opus-4 → input $15/M, output $75/M
	cost := EstimateCostUSD("claude-opus-4-6", 1000, 500)
	expected := (1000*15.0 + 500*75.0) / 1_000_000
	assert.InDelta(t, expected, cost, 1e-9)
}

func TestEstimateCostUSD_UnknownReturnsZero(t *testing.T) {
	cost := EstimateCostUSD("unknown-model", 1000, 500)
	assert.Equal(t, 0.0, cost)
}

func TestEstimateCostUSD_ZeroTokens(t *testing.T) {
	cost := EstimateCostUSD("claude-opus-4-6", 0, 0)
	assert.Equal(t, 0.0, cost)
}
