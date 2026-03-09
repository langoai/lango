package checks

import (
	"context"

	"github.com/langoai/lango/internal/config"
)

// ContractCheck validates smart contract configuration.
type ContractCheck struct{}

// Name returns the check name.
func (c *ContractCheck) Name() string {
	return "Smart Contracts"
}

// Run checks contract configuration validity.
func (c *ContractCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Payment.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Payment not enabled (contract interaction requires payment.enabled = true)",
		}
	}

	var issues []string
	status := StatusPass

	if cfg.Payment.Network.RPCURL == "" {
		issues = append(issues, "payment.network.rpcUrl is required for contract interaction")
		status = StatusFail
	}

	if cfg.Payment.Network.ChainID == 0 {
		issues = append(issues, "payment.network.chainId is required for contract interaction")
		status = StatusFail
	}

	if len(issues) == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Contract interaction configured (payment system provides RPC and chain ID)",
		}
	}

	message := "Contract issues:\n"
	for _, issue := range issues {
		message += "- " + issue + "\n"
	}
	return Result{Name: c.Name(), Status: status, Message: message}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *ContractCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
