package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	graphstore "github.com/langoai/lango/internal/graph"
)

func newImportCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import triples from a JSON file",
		Long: `Import triples into the knowledge graph from a JSON file.

The file should contain a JSON array of triple objects:
[
  {"Subject": "Alice", "Predicate": "knows", "Object": "Bob"},
  {"Subject": "Bob", "Predicate": "works_at", "Object": "Acme"}
]`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			var triples []graphstore.Triple
			if err := json.Unmarshal(data, &triples); err != nil {
				return fmt.Errorf("parse JSON: %w", err)
			}

			if len(triples) == 0 {
				fmt.Println("No triples to import.")
				return nil
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

			if err := store.AddTriples(context.Background(), triples); err != nil {
				return fmt.Errorf("import triples: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"imported": len(triples),
				})
			}

			fmt.Printf("Imported %d triples.\n", len(triples))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
