package graph

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newExportCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all triples from the knowledge graph",
		Long:  "Export all triples from the knowledge graph in JSON or CSV format.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if format != "json" && format != "csv" {
				return fmt.Errorf("--format must be json or csv")
			}

			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, err := initGraphStore(cfg)
			if err != nil {
				return err
			}
			defer store.Close()

			triples, err := store.AllTriples(context.Background())
			if err != nil {
				return fmt.Errorf("export triples: %w", err)
			}

			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(triples)

			case "csv":
				w := csv.NewWriter(os.Stdout)
				if err := w.Write([]string{"subject", "predicate", "object"}); err != nil {
					return fmt.Errorf("write csv header: %w", err)
				}
				for _, t := range triples {
					if err := w.Write([]string{t.Subject, t.Predicate, t.Object}); err != nil {
						return fmt.Errorf("write csv row: %w", err)
					}
				}
				w.Flush()
				return w.Error()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Output format: json or csv")

	return cmd
}
