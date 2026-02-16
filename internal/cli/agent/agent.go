package agent

import (
	"github.com/langowarny/lango/internal/config"
	"github.com/spf13/cobra"
)

// NewAgentCmd creates the agent command with lazy config loading.
func NewAgentCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Inspect agent mode and configuration",
	}

	cmd.AddCommand(newStatusCmd(cfgLoader))
	cmd.AddCommand(newListCmd(cfgLoader))

	return cmd
}
