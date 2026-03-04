package librarian

import (
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

// NewLibrarianCmd creates the librarian command with lazy bootstrap loading.
func NewLibrarianCmd(cfgLoader func() (*config.Config, error), bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "librarian",
		Short: "Inspect proactive knowledge librarian configuration",
	}

	cmd.AddCommand(newStatusCmd(cfgLoader))
	cmd.AddCommand(newInquiriesCmd(bootLoader))

	return cmd
}
