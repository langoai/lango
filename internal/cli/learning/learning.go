package learning

import (
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

// NewLearningCmd creates the learning command with lazy bootstrap loading.
func NewLearningCmd(cfgLoader func() (*config.Config, error), bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "learning",
		Short: "Inspect learning and knowledge configuration",
	}

	cmd.AddCommand(newStatusCmd(cfgLoader))
	cmd.AddCommand(newHistoryCmd(bootLoader))

	return cmd
}
