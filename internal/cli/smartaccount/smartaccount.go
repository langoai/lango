// Package smartaccount provides CLI commands for ERC-7579 smart account management.
package smartaccount

import (
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

// BootLoader returns bootstrap result for commands that need full app state.
type BootLoader func() (*bootstrap.Result, error)

// NewAccountCmd creates the "account" command with all subcommands.
func NewAccountCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "ERC-7579 Smart Account management",
		Long: `Manage Safe-based smart accounts with session keys, modules, and policies.

Examples:
  lango account info
  lango account deploy
  lango account session list
  lango account module list
  lango account policy show`,
	}

	cmd.AddCommand(deployCmd(bootLoader))
	cmd.AddCommand(infoCmd(bootLoader))
	cmd.AddCommand(sessionCmd(bootLoader))
	cmd.AddCommand(moduleCmd(bootLoader))
	cmd.AddCommand(policyCmd(bootLoader))
	cmd.AddCommand(paymasterCmd(bootLoader))

	return cmd
}
