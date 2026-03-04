package approval

import (
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

// NewApprovalCmd creates the approval command with lazy config loading.
func NewApprovalCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval",
		Short: "Inspect tool approval policy and providers",
	}

	cmd.AddCommand(newStatusCmd(cfgLoader))

	return cmd
}
