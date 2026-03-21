package provenance

import (
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newSessionCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "View session tree and hierarchy",
	}

	cmd.AddCommand(newSessionTreeCmd(bootLoader))
	cmd.AddCommand(newSessionListCmd(bootLoader))

	return cmd
}

func newSessionTreeCmd(_ func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "tree <session-key>",
		Short: "Show session hierarchy tree (not yet implemented)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("Session tree: not yet implemented (requires persistent session tree store)")
			return nil
		},
	}
}

func newSessionListCmd(_ func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all session nodes (not yet implemented)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("Session list: not yet implemented (requires persistent session tree store)")
			return nil
		},
	}
}
