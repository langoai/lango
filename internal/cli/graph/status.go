package graph

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show knowledge graph status",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			type statusOutput struct {
				Enabled      bool   `json:"enabled"`
				Backend      string `json:"backend"`
				DatabasePath string `json:"database_path"`
				TripleCount  int    `json:"triple_count"`
			}

			s := statusOutput{
				Enabled:      cfg.Graph.Enabled,
				Backend:      cfg.Graph.Backend,
				DatabasePath: cfg.Graph.DatabasePath,
			}

			if !cfg.Graph.Enabled {
				s.TripleCount = 0
				if output == "json" {
					return printJSON(cmd.OutOrStdout(), s)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Knowledge Graph Status")
				fmt.Fprintf(cmd.OutOrStdout(), "  Enabled:  %v\n", s.Enabled)
				return nil
			}

			store, err := initGraphStore(cfg)
			if err != nil {
				return err
			}
			defer store.Close()

			count, err := store.Count(context.Background())
			if err != nil {
				return fmt.Errorf("count triples: %w", err)
			}
			s.TripleCount = count

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), s)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Knowledge Graph Status")
			fmt.Fprintf(cmd.OutOrStdout(), "  Enabled:       %v\n", s.Enabled)
			fmt.Fprintf(cmd.OutOrStdout(), "  Backend:       %s\n", s.Backend)
			fmt.Fprintf(cmd.OutOrStdout(), "  Database Path: %s\n", s.DatabasePath)
			fmt.Fprintf(cmd.OutOrStdout(), "  Triples:       %d\n", s.TripleCount)

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	return cmd
}
