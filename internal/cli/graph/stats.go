package graph

import (
	"context"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatsCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "stats",
		Short:         "Show knowledge graph statistics",
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

			ctx := context.Background()

			total, err := store.Count(ctx)
			if err != nil {
				return fmt.Errorf("count triples: %w", err)
			}

			predicateStats, err := store.PredicateStats(ctx)
			if err != nil {
				return fmt.Errorf("predicate stats: %w", err)
			}

			type predicateEntry struct {
				Predicate string `json:"predicate"`
				Count     int    `json:"count"`
			}
			type statsOutput struct {
				TotalTriples   int              `json:"total_triples"`
				PredicateStats []predicateEntry `json:"predicate_stats"`
			}

			entries := make([]predicateEntry, 0, len(predicateStats))
			for p, c := range predicateStats {
				entries = append(entries, predicateEntry{Predicate: p, Count: c})
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Count > entries[j].Count
			})

			s := statsOutput{
				TotalTriples:   total,
				PredicateStats: entries,
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), s)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Knowledge Graph Statistics\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Total Triples: %d\n\n", s.TotalTriples)

			if len(entries) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No predicate data.")
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PREDICATE\tCOUNT")
			for _, e := range entries {
				fmt.Fprintf(w, "%s\t%d\n", e.Predicate, e.Count)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	return cmd
}
