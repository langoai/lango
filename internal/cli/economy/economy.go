// Package economy provides CLI commands for the P2P economy layer.
package economy

import (
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

// NewEconomyCmd creates the economy command group.
func NewEconomyCmd(
	cfgLoader func() (*config.Config, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "economy",
		Short: "Manage P2P economy (budget, risk, pricing, negotiation, escrow)",
		Long: `Manage the P2P economy layer for autonomous agent transactions.

Subcommands let you inspect budget allocations, assess risk, query pricing,
check negotiation sessions, and manage escrow agreements.

Examples:
  lango economy budget status --task-id=task-1
  lango economy risk status
  lango economy pricing status
  lango economy negotiate status
  lango economy escrow show`,
	}

	cmd.AddCommand(newBudgetCmd(cfgLoader))
	cmd.AddCommand(newRiskCmd(cfgLoader))
	cmd.AddCommand(newPricingCmd(cfgLoader))
	cmd.AddCommand(newNegotiateCmd(cfgLoader))
	cmd.AddCommand(newEscrowCmd(cfgLoader))

	return cmd
}
