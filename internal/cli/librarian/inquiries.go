package librarian

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/spf13/cobra"
)

func newInquiriesCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		output string
		limit  int
	)

	cmd := &cobra.Command{
		Use:           "inquiries",
		Short:         "List pending knowledge inquiries",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			if boot.Storage == nil {
				return fmt.Errorf("librarian storage unavailable")
			}
			entries, err := boot.Storage.PendingInquiries(cmd.Context(), limit)
			if err != nil {
				return fmt.Errorf("query inquiries: %w", err)
			}

			if output == "json" {
				type entry struct {
					ID       string `json:"id"`
					Topic    string `json:"topic"`
					Question string `json:"question"`
					Priority string `json:"priority"`
					Created  string `json:"created_at"`
				}

				out := make([]entry, 0, len(entries))
				for _, e := range entries {
					out = append(out, entry{
						ID:       e.ID,
						Topic:    e.Topic,
						Question: e.Question,
						Priority: e.Priority,
						Created:  e.Created.Format(time.RFC3339),
					})
				}

				return printJSON(cmd.OutOrStdout(), out)
			}

			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No pending inquiries.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Pending Inquiries (%d)\n", len(entries))
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tPRIORITY\tTOPIC\tQUESTION\tCREATED")
			for _, e := range entries {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					e.ID[:8],
					e.Priority,
					toolchain.Truncate(e.Topic, 22),
					toolchain.Truncate(e.Question, 37),
					e.Created.Format(time.DateTime),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of inquiries to show")

	return cmd
}
