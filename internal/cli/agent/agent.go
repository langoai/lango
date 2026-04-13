package agent

import (
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

// NewAgentCmd creates the agent command with lazy config loading.
// bootLoader is used by DB-requiring subcommands (trace, graph, trace metrics).
func NewAgentCmd(cfgLoader func() (*config.Config, error), bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Inspect agent mode and configuration",
	}

	cmd.AddCommand(newStatusCmd(cfgLoader))
	cmd.AddCommand(newListCmd(cfgLoader))
	cmd.AddCommand(newToolsCmd(cfgLoader))
	cmd.AddCommand(newHooksCmd(cfgLoader))
	cmd.AddCommand(newTraceCmd(bootLoader))
	cmd.AddCommand(newGraphCmd(bootLoader))
	cmd.AddCommand(newTraceMetricsCmd(bootLoader))

	return cmd
}
