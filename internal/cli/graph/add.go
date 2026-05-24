package graph

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	graphstore "github.com/langoai/lango/internal/graph"
)

func newAddCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		subject   string
		predicate string
		object    string
		output    string
	)

	cmd := &cobra.Command{
		Use:           "add",
		Short:         "Add a triple to the knowledge graph",
		Long:          "Add a subject-predicate-object triple to the knowledge graph store.",
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

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), triple)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added triple: (%s) -[%s]-> (%s)\n", subject, predicate, object)
			return nil
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Subject of the triple")
	cmd.Flags().StringVar(&predicate, "predicate", "", "Predicate (relationship) of the triple")
	cmd.Flags().StringVar(&object, "object", "", "Object of the triple")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	_ = cmd.MarkFlagRequired("subject")
	_ = cmd.MarkFlagRequired("predicate")
	_ = cmd.MarkFlagRequired("object")

	return cmd
}
