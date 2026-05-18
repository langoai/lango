package settings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestWave29EconomyFormFormatsBudgetFields(t *testing.T) {
	hardLimit := false
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Budget.DefaultMax = "42.50"
	cfg.Economy.Budget.HardLimit = &hardLimit
	cfg.Economy.Budget.AlertThresholds = []float64{0.5, 0.8, 0.95}

	form := NewEconomyForm(cfg)

	assert.Equal(t, "Economy Configuration", form.Title)
	assert.True(t, fieldByKey(form, "economy_enabled").Checked)
	assert.Equal(t, "42.50", fieldByKey(form, "economy_budget_default_max").Value)
	assert.False(t, fieldByKey(form, "economy_budget_hard_limit").Checked)
	assert.Equal(t, "0.5,0.8,0.95", fieldByKey(form, "economy_budget_alert_thresholds").Value)
	assert.Empty(t, formatFloatSlice(nil))
}

func TestWave29EconomyRiskFormValidatesTrustScores(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Risk.EscrowThreshold = "12.00"
	cfg.Economy.Risk.HighTrustScore = 0.8
	cfg.Economy.Risk.MediumTrustScore = 0.5

	form := NewEconomyRiskForm(cfg)

	assert.Equal(t, "12.00", fieldByKey(form, "economy_risk_escrow_threshold").Value)
	highTrust := fieldByKey(form, "economy_risk_high_trust")
	mediumTrust := fieldByKey(form, "economy_risk_medium_trust")
	assert.Equal(t, "0.8", highTrust.Value)
	assert.Equal(t, "0.5", mediumTrust.Value)
	require.NoError(t, highTrust.Validate("1.0"))
	require.NoError(t, mediumTrust.Validate("0"))
	assert.EqualError(t, highTrust.Validate("NaN-ish"), "must be a number")
	assert.EqualError(t, mediumTrust.Validate("1.1"), "must be between 0.0 and 1.0")
}

func TestWave29EconomyNegotiationFormValidatesRoundsAndDiscount(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Negotiate.Enabled = true
	cfg.Economy.Negotiate.MaxRounds = 7
	cfg.Economy.Negotiate.Timeout = 9 * time.Minute
	cfg.Economy.Negotiate.AutoNegotiate = true
	cfg.Economy.Negotiate.MaxDiscount = 0.25

	form := NewEconomyNegotiationForm(cfg)

	assert.True(t, fieldByKey(form, "economy_negotiate_enabled").Checked)
	assert.Equal(t, "7", fieldByKey(form, "economy_negotiate_max_rounds").Value)
	assert.Equal(t, "9m0s", fieldByKey(form, "economy_negotiate_timeout").Value)
	assert.True(t, fieldByKey(form, "economy_negotiate_auto").Checked)
	assert.Equal(t, "0.25", fieldByKey(form, "economy_negotiate_max_discount").Value)

	maxRounds := fieldByKey(form, "economy_negotiate_max_rounds")
	maxDiscount := fieldByKey(form, "economy_negotiate_max_discount")
	require.NoError(t, maxRounds.Validate("1"))
	assert.EqualError(t, maxRounds.Validate("0"), "must be a positive integer")
	assert.EqualError(t, maxRounds.Validate("abc"), "must be a positive integer")
	require.NoError(t, maxDiscount.Validate("0.75"))
	assert.EqualError(t, maxDiscount.Validate("-0.1"), "must be between 0.0 and 1.0")
	assert.EqualError(t, maxDiscount.Validate("abc"), "must be a number")
}

func TestWave29EconomyEscrowFormsExposeDurationsAndValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Escrow.Enabled = true
	cfg.Economy.Escrow.DefaultTimeout = 48 * time.Hour
	cfg.Economy.Escrow.MaxMilestones = 4
	cfg.Economy.Escrow.AutoRelease = true
	cfg.Economy.Escrow.DisputeWindow = 2 * time.Hour

	form := NewEconomyEscrowForm(cfg)

	assert.True(t, fieldByKey(form, "economy_escrow_enabled").Checked)
	assert.Equal(t, "48h0m0s", fieldByKey(form, "economy_escrow_default_timeout").Value)
	assert.Equal(t, "4", fieldByKey(form, "economy_escrow_max_milestones").Value)
	assert.True(t, fieldByKey(form, "economy_escrow_auto_release").Checked)
	assert.Equal(t, "2h0m0s", fieldByKey(form, "economy_escrow_dispute_window").Value)

	maxMilestones := fieldByKey(form, "economy_escrow_max_milestones")
	require.NoError(t, maxMilestones.Validate("1"))
	assert.EqualError(t, maxMilestones.Validate("0"), "must be a positive integer")
}

func TestWave29EconomyOnChainFormValidatesModeDepthAndRetries(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Escrow.OnChain.Enabled = true
	cfg.Economy.Escrow.OnChain.Mode = "vault"
	cfg.Economy.Escrow.OnChain.HubAddress = "0xhub"
	cfg.Economy.Escrow.OnChain.VaultFactoryAddress = "0xfactory"
	cfg.Economy.Escrow.OnChain.VaultImplementation = "0ximpl"
	cfg.Economy.Escrow.OnChain.ArbitratorAddress = "0xarb"
	cfg.Economy.Escrow.OnChain.TokenAddress = "0xtoken"
	cfg.Economy.Escrow.OnChain.PollInterval = 30 * time.Second
	cfg.Economy.Escrow.OnChain.ConfirmationDepth = 5
	cfg.Economy.Escrow.Settlement.ReceiptTimeout = 3 * time.Minute
	cfg.Economy.Escrow.Settlement.MaxRetries = 2

	form := NewEconomyEscrowOnChainForm(cfg)

	assert.True(t, fieldByKey(form, "economy_escrow_onchain_enabled").Checked)
	assert.Equal(t, "vault", fieldByKey(form, "economy_escrow_onchain_mode").Value)
	assert.Equal(t, "0xhub", fieldByKey(form, "economy_escrow_onchain_hub_address").Value)
	assert.Equal(t, "0xfactory", fieldByKey(form, "economy_escrow_onchain_vault_factory").Value)
	assert.Equal(t, "0ximpl", fieldByKey(form, "economy_escrow_onchain_vault_impl").Value)
	assert.Equal(t, "0xarb", fieldByKey(form, "economy_escrow_onchain_arbitrator").Value)
	assert.Equal(t, "0xtoken", fieldByKey(form, "economy_escrow_onchain_token").Value)
	assert.Equal(t, "30s", fieldByKey(form, "economy_escrow_onchain_poll_interval").Value)
	assert.Equal(t, "5", fieldByKey(form, "economy_escrow_onchain_confirmation_depth").Value)
	assert.Equal(t, "3m0s", fieldByKey(form, "economy_escrow_settlement_receipt_timeout").Value)
	assert.Equal(t, "2", fieldByKey(form, "economy_escrow_settlement_max_retries").Value)

	mode := fieldByKey(form, "economy_escrow_onchain_mode")
	depth := fieldByKey(form, "economy_escrow_onchain_confirmation_depth")
	retries := fieldByKey(form, "economy_escrow_settlement_max_retries")
	require.NoError(t, mode.Validate("hub"))
	require.NoError(t, mode.Validate("vault"))
	assert.EqualError(t, mode.Validate("direct"), "must be 'hub' or 'vault'")
	require.NoError(t, depth.Validate("0"))
	assert.EqualError(t, depth.Validate("-1"), "must be a non-negative integer")
	require.NoError(t, retries.Validate("0"))
	assert.EqualError(t, retries.Validate("-1"), "must be a non-negative integer")
}

func TestWave29EconomyPricingFormValidatesDiscounts(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Pricing.Enabled = true
	cfg.Economy.Pricing.TrustDiscount = 0.1
	cfg.Economy.Pricing.VolumeDiscount = 0.05
	cfg.Economy.Pricing.MinPrice = "0.02"

	form := NewEconomyPricingForm(cfg)

	assert.True(t, fieldByKey(form, "economy_pricing_enabled").Checked)
	assert.Equal(t, "0.10", fieldByKey(form, "economy_pricing_trust_discount").Value)
	assert.Equal(t, "0.05", fieldByKey(form, "economy_pricing_volume_discount").Value)
	assert.Equal(t, "0.02", fieldByKey(form, "economy_pricing_min_price").Value)

	trustDiscount := fieldByKey(form, "economy_pricing_trust_discount")
	volumeDiscount := fieldByKey(form, "economy_pricing_volume_discount")
	require.NoError(t, trustDiscount.Validate("0"))
	require.NoError(t, volumeDiscount.Validate("1"))
	assert.EqualError(t, trustDiscount.Validate("abc"), "must be a number")
	assert.EqualError(t, volumeDiscount.Validate("1.01"), "must be between 0.0 and 1.0")
}
