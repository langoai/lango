package graph

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/langoai/lango/internal/config"
	graphstore "github.com/langoai/lango/internal/graph"
	"github.com/spf13/cobra"
)

func newQueryCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		subject   string
		predicate string
		object    string
		limit     int
		output    string
	)

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query triples from the knowledge graph",
		Long: `Query triples by subject, object, or subject+predicate.

At least one of --subject or --object is required.
The --predicate flag can only be used together with --subject.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			if subject == "" && object == "" {
				return fmt.Errorf("at least one of --subject or --object is required")
			}
			if predicate != "" && subject == "" {
				return fmt.Errorf("--predicate requires --subject")
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

			ctx := context.Background()
			var triples []graphstore.Triple

			switch {
			case subject != "" && predicate != "":
				triples, err = store.QueryBySubjectPredicate(ctx, subject, predicate)
			case subject != "":
				triples, err = store.QueryBySubject(ctx, subject)
			case object != "":
				triples, err = store.QueryByObject(ctx, object)
			}
			if err != nil {
				return fmt.Errorf("query triples: %w", err)
			}

			if limit > 0 && len(triples) > limit {
				triples = triples[:limit]
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), triples)
			}

			if len(triples) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No triples found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SUBJECT\tPREDICATE\tOBJECT")
			for _, t := range triples {
				fmt.Fprintf(w, "%s\t%s\t%s\n", t.Subject, t.Predicate, t.Object)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Filter by subject")
	cmd.Flags().StringVar(&predicate, "predicate", "", "Filter by predicate (requires --subject)")
	cmd.Flags().StringVar(&object, "object", "", "Filter by object")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of results (0 = unlimited)")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	return cmd
}
