package config

import (
	"testing"
	"time"
)

func TestEconomyConfig_ZeroValues(t *testing.T) {
	var cfg EconomyConfig

	if cfg.Enabled {
		t.Error("Enabled should default to false")
	}

	if cfg.Budget.DefaultMax != "" {
		t.Error("Budget.DefaultMax should default to empty string")
	}
	if cfg.Budget.AlertThresholds != nil {
		t.Error("Budget.AlertThresholds should default to nil")
	}
	if cfg.Budget.HardLimit != nil {
		t.Error("Budget.HardLimit should default to nil (use-default sentinel)")
	}

	if cfg.Risk.EscrowThreshold != "" {
		t.Error("Risk.EscrowThreshold should default to empty string")
	}
	if cfg.Risk.HighTrustScore != 0 {
		t.Error("Risk.HighTrustScore should default to 0")
	}
	if cfg.Risk.MediumTrustScore != 0 {
		t.Error("Risk.MediumTrustScore should default to 0")
	}

	if cfg.Negotiate.Enabled {
		t.Error("Negotiate.Enabled should default to false")
	}
	if cfg.Negotiate.MaxRounds != 0 {
		t.Error("Negotiate.MaxRounds should default to 0")
	}
	if cfg.Negotiate.Timeout != 0 {
		t.Error("Negotiate.Timeout should default to 0")
	}

	if cfg.Escrow.Enabled {
		t.Error("Escrow.Enabled should default to false")
	}
	if cfg.Escrow.DefaultTimeout != 0 {
		t.Error("Escrow.DefaultTimeout should default to 0")
	}
	if cfg.Escrow.MaxMilestones != 0 {
		t.Error("Escrow.MaxMilestones should default to 0")
	}

	if cfg.Pricing.Enabled {
		t.Error("Pricing.Enabled should default to false")
	}
	if cfg.Pricing.MinPrice != "" {
		t.Error("Pricing.MinPrice should default to empty string")
	}
}

func TestBudgetConfig_HardLimitPointer(t *testing.T) {
	hardLimit := true
	cfg := BudgetConfig{
		DefaultMax:      "10.00",
		AlertThresholds: []float64{0.5, 0.8, 0.95},
		HardLimit:       &hardLimit,
	}

	if cfg.DefaultMax != "10.00" {
		t.Errorf("DefaultMax = %q, want %q", cfg.DefaultMax, "10.00")
	}
	if len(cfg.AlertThresholds) != 3 {
		t.Errorf("AlertThresholds length = %d, want 3", len(cfg.AlertThresholds))
	}
	if cfg.HardLimit == nil || !*cfg.HardLimit {
		t.Error("HardLimit should be non-nil and true")
	}
}

func TestNegotiationConfig_Timeout(t *testing.T) {
	cfg := NegotiationConfig{
		Enabled:       true,
		MaxRounds:     5,
		Timeout:       5 * time.Minute,
		AutoNegotiate: true,
		MaxDiscount:   0.2,
	}

	if cfg.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want 5m", cfg.Timeout)
	}
	if cfg.MaxDiscount != 0.2 {
		t.Errorf("MaxDiscount = %f, want 0.2", cfg.MaxDiscount)
	}
}

func TestEscrowConfig_Durations(t *testing.T) {
	cfg := EscrowConfig{
		Enabled:        true,
		DefaultTimeout: 24 * time.Hour,
		MaxMilestones:  10,
		AutoRelease:    true,
		DisputeWindow:  time.Hour,
	}

	if cfg.DefaultTimeout != 24*time.Hour {
		t.Errorf("DefaultTimeout = %v, want 24h", cfg.DefaultTimeout)
	}
	if cfg.DisputeWindow != time.Hour {
		t.Errorf("DisputeWindow = %v, want 1h", cfg.DisputeWindow)
	}
}

func TestConfigHasEconomyField(t *testing.T) {
	var cfg Config
	cfg.Economy.Enabled = true
	if !cfg.Economy.Enabled {
		t.Error("Config.Economy.Enabled should be settable")
	}
}
