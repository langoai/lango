package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatsCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show knowledge graph statistics",
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(s)
			}

			fmt.Printf("Knowledge Graph Statistics\n")
			fmt.Printf("  Total Triples: %d\n\n", s.TotalTriples)

			if len(entries) == 0 {
				fmt.Println("No predicate data.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PREDICATE\tCOUNT")
			for _, e := range entries {
				fmt.Fprintf(w, "%s\t%d\n", e.Predicate, e.Count)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
