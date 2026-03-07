package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewEconomyForm creates the Economy configuration form.
func NewEconomyForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Economy Configuration")

	form.AddField(&tuicore.Field{
		Key: "economy_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Economy.Enabled,
		Description: "Enable the P2P economy layer (budget, risk, pricing, negotiation, escrow)",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_budget_default_max", Label: "Default Budget Max (USDC)", Type: tuicore.InputText,
		Value:       cfg.Economy.Budget.DefaultMax,
		Placeholder: "10.00",
		Description: "Default maximum budget per task in USDC",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_budget_hard_limit", Label: "Hard Limit", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.Economy.Budget.HardLimit, true),
		Description: "Enforce budget as a hard cap (reject overspend)",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_budget_alert_thresholds", Label: "Alert Thresholds", Type: tuicore.InputText,
		Value:       formatFloatSlice(cfg.Economy.Budget.AlertThresholds),
		Placeholder: "0.5,0.8,0.95 (comma-separated percentages)",
		Description: "Budget usage percentages that trigger alerts",
	})

	return &form
}

// NewEconomyRiskForm creates the Economy Risk configuration form.
func NewEconomyRiskForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Economy Risk Configuration")

	form.AddField(&tuicore.Field{
		Key: "economy_risk_escrow_threshold", Label: "Escrow Threshold (USDC)", Type: tuicore.InputText,
		Value:       cfg.Economy.Risk.EscrowThreshold,
		Placeholder: "5.00",
		Description: "USDC amount above which escrow is forced",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_risk_high_trust", Label: "High Trust Score", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.1f", cfg.Economy.Risk.HighTrustScore),
		Placeholder: "0.8 (0.0 to 1.0)",
		Description: "Minimum trust score for DirectPay strategy",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "economy_risk_medium_trust", Label: "Medium Trust Score", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.1f", cfg.Economy.Risk.MediumTrustScore),
		Placeholder: "0.5 (0.0 to 1.0)",
		Description: "Minimum trust score for non-ZK strategies",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
	})

	return &form
}

// NewEconomyNegotiationForm creates the Economy Negotiation configuration form.
func NewEconomyNegotiationForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Economy Negotiation Configuration")

	form.AddField(&tuicore.Field{
		Key: "economy_negotiate_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Economy.Negotiate.Enabled,
		Description: "Enable the P2P negotiation protocol",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_negotiate_max_rounds", Label: "Max Rounds", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Economy.Negotiate.MaxRounds),
		Placeholder: "5",
		Description: "Maximum number of counter-offers per negotiation",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "economy_negotiate_timeout", Label: "Timeout", Type: tuicore.InputText,
		Value:       cfg.Economy.Negotiate.Timeout.String(),
		Placeholder: "5m",
		Description: "Negotiation session timeout duration",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_negotiate_auto", Label: "Auto-Negotiate", Type: tuicore.InputBool,
		Checked:     cfg.Economy.Negotiate.AutoNegotiate,
		Description: "Automatically generate counter-offers",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_negotiate_max_discount", Label: "Max Discount", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.2f", cfg.Economy.Negotiate.MaxDiscount),
		Placeholder: "0.20 (0.0 to 1.0)",
		Description: "Maximum discount percentage for auto-negotiation",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
	})

	return &form
}

// NewEconomyEscrowForm creates the Economy Escrow configuration form.
func NewEconomyEscrowForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Economy Escrow Configuration")

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Economy.Escrow.Enabled,
		Description: "Enable the milestone-based escrow service",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_default_timeout", Label: "Default Timeout", Type: tuicore.InputText,
		Value:       cfg.Economy.Escrow.DefaultTimeout.String(),
		Placeholder: "24h",
		Description: "Escrow expiration timeout",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_max_milestones", Label: "Max Milestones", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Economy.Escrow.MaxMilestones),
		Placeholder: "10",
		Description: "Maximum milestones per escrow",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_auto_release", Label: "Auto-Release", Type: tuicore.InputBool,
		Checked:     cfg.Economy.Escrow.AutoRelease,
		Description: "Automatically release funds when all milestones are met",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_dispute_window", Label: "Dispute Window", Type: tuicore.InputText,
		Value:       cfg.Economy.Escrow.DisputeWindow.String(),
		Placeholder: "1h",
		Description: "Time window for raising disputes after completion",
	})

	return &form
}

// NewEconomyEscrowOnChainForm creates the on-chain escrow configuration form.
func NewEconomyEscrowOnChainForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("On-Chain Escrow Configuration")
	oc := cfg.Economy.Escrow.OnChain
	st := cfg.Economy.Escrow.Settlement

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_onchain_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     oc.Enabled,
		Description: "Enable on-chain escrow mode",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_onchain_mode", Label: "Mode", Type: tuicore.InputText,
		Value:       oc.Mode,
		Placeholder: "hub (hub or vault)",
		Description: "On-chain escrow pattern: hub (single contract) or vault (per-deal clone)",
		Validate: func(s string) error {
			if s != "hub" && s != "vault" {
				return fmt.Errorf("must be 'hub' or 'vault'")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_onchain_hub_address", Label: "Hub Address", Type: tuicore.InputText,
		Value:       oc.HubAddress,
		Placeholder: "0x...",
		Description: "Deployed LangoEscrowHub contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_onchain_vault_factory", Label: "Vault Factory Address", Type: tuicore.InputText,
		Value:       oc.VaultFactoryAddress,
		Placeholder: "0x...",
		Description: "Deployed LangoVaultFactory contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_onchain_vault_impl", Label: "Vault Implementation", Type: tuicore.InputText,
		Value:       oc.VaultImplementation,
		Placeholder: "0x...",
		Description: "LangoVault implementation address for cloning",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_onchain_arbitrator", Label: "Arbitrator Address", Type: tuicore.InputText,
		Value:       oc.ArbitratorAddress,
		Placeholder: "0x...",
		Description: "Dispute arbitrator address",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_onchain_token", Label: "Token Address", Type: tuicore.InputText,
		Value:       oc.TokenAddress,
		Placeholder: "0x...",
		Description: "ERC-20 token (USDC) contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_onchain_poll_interval", Label: "Poll Interval", Type: tuicore.InputText,
		Value:       oc.PollInterval.String(),
		Placeholder: "15s",
		Description: "Event monitor polling interval",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_settlement_receipt_timeout", Label: "Receipt Timeout", Type: tuicore.InputText,
		Value:       st.ReceiptTimeout.String(),
		Placeholder: "2m",
		Description: "Max wait for on-chain receipt confirmation",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_escrow_settlement_max_retries", Label: "Max Retries", Type: tuicore.InputInt,
		Value:       strconv.Itoa(st.MaxRetries),
		Placeholder: "3",
		Description: "Max transaction submission retries",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	return &form
}

// NewEconomyPricingForm creates the Economy Dynamic Pricing configuration form.
func NewEconomyPricingForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Economy Pricing Configuration")

	form.AddField(&tuicore.Field{
		Key: "economy_pricing_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Economy.Pricing.Enabled,
		Description: "Enable dynamic pricing adjustments",
	})

	form.AddField(&tuicore.Field{
		Key: "economy_pricing_trust_discount", Label: "Trust Discount", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.2f", cfg.Economy.Pricing.TrustDiscount),
		Placeholder: "0.10 (0.0 to 1.0)",
		Description: "Maximum discount for high-trust peers",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "economy_pricing_volume_discount", Label: "Volume Discount", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.2f", cfg.Economy.Pricing.VolumeDiscount),
		Placeholder: "0.05 (0.0 to 1.0)",
		Description: "Maximum discount for high-volume peers",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "economy_pricing_min_price", Label: "Min Price (USDC)", Type: tuicore.InputText,
		Value:       cfg.Economy.Pricing.MinPrice,
		Placeholder: "0.01",
		Description: "Minimum price floor in USDC",
	})

	return &form
}

// formatFloatSlice formats a float64 slice as a comma-separated string.
func formatFloatSlice(vals []float64) string {
	if len(vals) == 0 {
		return ""
	}
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return strings.Join(parts, ",")
}
