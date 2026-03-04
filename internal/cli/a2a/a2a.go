package a2a

import (
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

// NewA2ACmd creates the a2a command with lazy config loading.
func NewA2ACmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "a2a",
		Short: "Inspect A2A (Agent-to-Agent) protocol configuration",
	}

	cmd.AddCommand(newCardCmd(cfgLoader))
	cmd.AddCommand(newCheckCmd())

	return cmd
}
