// Package contract provides CLI commands for smart contract interaction.
package contract

import (
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

// NewContractCmd creates the contract command group.
func NewContractCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "Interact with smart contracts",
		Long: `Read and write to smart contracts on EVM chains.

Examples:
  lango contract read --address 0x... --abi ./erc20.json --method balanceOf --args 0x...
  lango contract call --address 0x... --abi ./erc20.json --method transfer --args 0x...,1000000
  lango contract abi load --address 0x... --file ./erc20.json`,
	}

	cmd.AddCommand(newReadCmd(cfgLoader))
	cmd.AddCommand(newCallCmd(cfgLoader))
	cmd.AddCommand(newABICmd(cfgLoader))

	return cmd
}
