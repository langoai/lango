package learning

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/spf13/cobra"
)

func newHistoryCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		jsonOutput bool
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recent learning entries and corrections",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			if boot.Storage == nil {
				return fmt.Errorf("learning storage unavailable")
			}
			entries, err := boot.Storage.LearningHistory(cmd.Context(), limit)
			if err != nil {
				return fmt.Errorf("query learnings: %w", err)
			}

			if jsonOutput {
				return printHistoryJSON(entries)
			}
			return printHistoryTable(entries)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of entries to show")

	return cmd
}

func printHistoryJSON(entries []storage.LearningHistoryRecord) error {
	type entry struct {
		ID         string  `json:"id"`
		Trigger    string  `json:"trigger"`
		Category   string  `json:"category"`
		Diagnosis  string  `json:"diagnosis"`
		Fix        string  `json:"fix,omitempty"`
		Confidence float64 `json:"confidence"`
		CreatedAt  string  `json:"created_at"`
	}

	out := make([]entry, 0, len(entries))
	for _, e := range entries {
		out = append(out, entry{
			ID:         e.ID,
			Trigger:    e.Trigger,
			Category:   e.Category,
			Diagnosis:  e.Diagnosis,
			Fix:        e.Fix,
			Confidence: e.Confidence,
			CreatedAt:  e.CreatedAt.Format(time.RFC3339),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printHistoryTable(entries []storage.LearningHistoryRecord) error {
	if len(entries) == 0 {
		fmt.Println("No learning entries found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCATEGORY\tTRIGGER\tCONFIDENCE\tCREATED")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\t%s\n",
			e.ID[:8],
			e.Category,
			toolchain.Truncate(e.Trigger, 37),
			e.Confidence,
			e.CreatedAt.Format(time.DateTime),
		)
	}
	return w.Flush()
}
