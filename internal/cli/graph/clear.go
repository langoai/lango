package graph

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
		Use:   "clear",
		Short: "Clear all triples from the knowledge graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, err := initGraphStore(cfg)
			if err != nil {
				return err
			}
			defer store.Close()

			if !force {
				fmt.Fprintln(cmd.OutOrStdout(), "This will delete all triples from the knowledge graph.")
				ok, err := prompt.ConfirmDenyOnEOFIO(cmd.InOrStdin(), cmd.OutOrStdout(), "Continue?")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			if err := store.ClearAll(context.Background()); err != nil {
				return fmt.Errorf("clear graph: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Cleared all triples from the knowledge graph.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
