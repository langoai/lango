package checks

import (
	"context"
	"fmt"
	"math/big"

	"github.com/langoai/lango/internal/config"
)

// EconomyCheck validates economy layer configuration.
type EconomyCheck struct{}

// Name returns the check name.
func (c *EconomyCheck) Name() string {
	return "Economy Layer"
}

// Run checks economy configuration validity.
func (c *EconomyCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Economy.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Economy layer not enabled (economy.enabled = false)",
		}
	}

	var issues []string
	status := StatusPass

	// Validate budget.defaultMax is parseable as a decimal.
	if cfg.Economy.Budget.DefaultMax != "" {
		if _, _, err := new(big.Float).Parse(cfg.Economy.Budget.DefaultMax, 10); err != nil {
			issues = append(issues, fmt.Sprintf("budget.defaultMax %q is not a valid decimal", cfg.Economy.Budget.DefaultMax))
			status = StatusFail
		}
	}

	// Validate risk score ordering.
	if cfg.Economy.Risk.HighTrustScore > 0 && cfg.Economy.Risk.MediumTrustScore > 0 {
		if cfg.Economy.Risk.HighTrustScore <= cfg.Economy.Risk.MediumTrustScore {
			issues = append(issues, fmt.Sprintf("risk.highTrustScore (%.2f) should be greater than mediumTrustScore (%.2f)",
				cfg.Economy.Risk.HighTrustScore, cfg.Economy.Risk.MediumTrustScore))
			if status < StatusWarn {
				status = StatusWarn
			}
		}
	}

	// Validate escrow.maxMilestones.
	if cfg.Economy.Escrow.Enabled && cfg.Economy.Escrow.MaxMilestones <= 0 {
		issues = append(issues, "escrow.maxMilestones should be positive when escrow is enabled")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	// Validate negotiate.maxRounds.
	if cfg.Economy.Negotiate.Enabled && cfg.Economy.Negotiate.MaxRounds <= 0 {
		issues = append(issues, "negotiate.maxRounds should be positive when negotiation is enabled")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	// Validate pricing.minPrice.
	if cfg.Economy.Pricing.Enabled && cfg.Economy.Pricing.MinPrice != "" {
		if _, _, err := new(big.Float).Parse(cfg.Economy.Pricing.MinPrice, 10); err != nil {
			issues = append(issues, fmt.Sprintf("pricing.minPrice %q is not a valid decimal", cfg.Economy.Pricing.MinPrice))
			status = StatusFail
		}
	}

	if len(issues) == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Economy layer configured",
		}
	}

	message := "Economy layer issues:\n"
	for _, issue := range issues {
		message += fmt.Sprintf("- %s\n", issue)
	}
	return Result{Name: c.Name(), Status: status, Message: message}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *EconomyCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
