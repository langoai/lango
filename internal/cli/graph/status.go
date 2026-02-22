package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show knowledge graph status",
		RunE: func(cmd *cobra.Command, args []string) error {
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
				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(s)
				}
				fmt.Println("Knowledge Graph Status")
				fmt.Printf("  Enabled:  %v\n", s.Enabled)
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(s)
			}

			fmt.Println("Knowledge Graph Status")
			fmt.Printf("  Enabled:       %v\n", s.Enabled)
			fmt.Printf("  Backend:       %s\n", s.Backend)
			fmt.Printf("  Database Path: %s\n", s.DatabasePath)
			fmt.Printf("  Triples:       %d\n", s.TripleCount)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
