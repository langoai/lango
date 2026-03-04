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

func newAddCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		subject    string
		predicate  string
		object     string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a triple to the knowledge graph",
		Long:  "Add a subject-predicate-object triple to the knowledge graph store.",
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

			triple := graphstore.Triple{
				Subject:   subject,
				Predicate: predicate,
				Object:    object,
			}

			if err := store.AddTriple(context.Background(), triple); err != nil {
				return fmt.Errorf("add triple: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(triple)
			}

			fmt.Printf("Added triple: (%s) -[%s]-> (%s)\n", subject, predicate, object)
			return nil
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Subject of the triple")
	cmd.Flags().StringVar(&predicate, "predicate", "", "Predicate (relationship) of the triple")
	cmd.Flags().StringVar(&object, "object", "", "Object of the triple")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("subject")
	_ = cmd.MarkFlagRequired("predicate")
	_ = cmd.MarkFlagRequired("object")

	return cmd
}
