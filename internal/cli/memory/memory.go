package memory

import (
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

// NewMemoryCmd creates the memory command with lazy config loading.
func NewMemoryCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Manage observational memory",
	}

	cmd.AddCommand(newListCmd(cfgLoader))
	cmd.AddCommand(newStatusCmd(cfgLoader))
	cmd.AddCommand(newClearCmd(cfgLoader))

	return cmd
}
