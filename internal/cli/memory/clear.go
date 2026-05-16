package memory

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newClearCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "clear <session-key>",
		Short: "Clear all observations and reflections for a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionKey := args[0]

			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, memStore, cleanup, err := initMemoryStore(cfg)
			if err != nil {
				return err
			}
			_ = store
			defer cleanup()

			if !force {
				fmt.Fprintf(cmd.OutOrStdout(), "This will delete all observations and reflections for session '%s'.\n", sessionKey)
				ok, err := prompt.ConfirmDenyOnEOFIO(cmd.InOrStdin(), cmd.OutOrStdout(), "Continue?")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			ctx := context.Background()

			if err := memStore.DeleteObservationsBySession(ctx, sessionKey); err != nil {
				return fmt.Errorf("delete observations: %w", err)
			}

			if err := memStore.DeleteReflectionsBySession(ctx, sessionKey); err != nil {
				return fmt.Errorf("delete reflections: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cleared all memory entries for session '%s'.\n", sessionKey)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
