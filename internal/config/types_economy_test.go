package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEconomyConfig_ZeroValues(t *testing.T) {
	t.Parallel()

	var cfg EconomyConfig

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.Budget.DefaultMax)
	assert.Nil(t, cfg.Budget.AlertThresholds)
	assert.Nil(t, cfg.Budget.HardLimit)
	assert.Empty(t, cfg.Risk.EscrowThreshold)
	assert.Zero(t, cfg.Risk.HighTrustScore)
	assert.Zero(t, cfg.Risk.MediumTrustScore)
	assert.False(t, cfg.Negotiate.Enabled)
	assert.Zero(t, cfg.Negotiate.MaxRounds)
	assert.Zero(t, cfg.Negotiate.Timeout)
	assert.False(t, cfg.Escrow.Enabled)
	assert.Zero(t, cfg.Escrow.DefaultTimeout)
	assert.Zero(t, cfg.Escrow.MaxMilestones)
	assert.False(t, cfg.Pricing.Enabled)
	assert.Empty(t, cfg.Pricing.MinPrice)
}

func TestBudgetConfig_HardLimitPointer(t *testing.T) {
	t.Parallel()

	hardLimit := true
	cfg := BudgetConfig{
		DefaultMax:      "10.00",
		AlertThresholds: []float64{0.5, 0.8, 0.95},
		HardLimit:       &hardLimit,
	}

	assert.Equal(t, "10.00", cfg.DefaultMax)
	assert.Len(t, cfg.AlertThresholds, 3)
	require.NotNil(t, cfg.HardLimit)
	assert.True(t, *cfg.HardLimit)
}

func TestNegotiationConfig_Timeout(t *testing.T) {
	t.Parallel()

	cfg := NegotiationConfig{
		Enabled:       true,
		MaxRounds:     5,
		Timeout:       5 * time.Minute,
		AutoNegotiate: true,
		MaxDiscount:   0.2,
	}

	assert.Equal(t, 5*time.Minute, cfg.Timeout)
	assert.InDelta(t, 0.2, cfg.MaxDiscount, 1e-9)
}

func TestEscrowConfig_Durations(t *testing.T) {
	t.Parallel()

	cfg := EscrowConfig{
		Enabled:        true,
		DefaultTimeout: 24 * time.Hour,
		MaxMilestones:  10,
		AutoRelease:    true,
		DisputeWindow:  time.Hour,
	}

	assert.Equal(t, 24*time.Hour, cfg.DefaultTimeout)
	assert.Equal(t, time.Hour, cfg.DisputeWindow)
}

func TestConfigHasEconomyField(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Economy.Enabled = true
	assert.True(t, cfg.Economy.Enabled)
}
