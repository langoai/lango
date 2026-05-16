package metrics

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/clihttp"
)

func newPolicyCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "policy",
		Short:         "Policy decision statistics",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `View policy decision metrics including block and observe counts
with per-reason breakdowns.

Examples:
  lango metrics policy                  # Table summary
  lango metrics policy --output json    # JSON output`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format, err := getOutputFormat(cmd)
			if err != nil {
				return err
			}

			var data struct {
				Blocks   int64            `json:"blocks"`
				Observes int64            `json:"observes"`
				ByReason map[string]int64 `json:"byReason"`
			}
			if err := clihttp.FetchJSON(addr, "/metrics/policy", &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(cmd.OutOrStdout(), data)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "=== Policy Decisions ===")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintf(cmd.OutOrStdout(), "Blocks:    %d\n", data.Blocks)
			fmt.Fprintf(cmd.OutOrStdout(), "Observes:  %d\n", data.Observes)

			if len(data.ByReason) > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), "--- By Reason ---")
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "REASON\tCOUNT")

				reasons := make([]string, 0, len(data.ByReason))
				for r := range data.ByReason {
					reasons = append(reasons, r)
				}
				sort.Strings(reasons)

				for _, r := range reasons {
					fmt.Fprintf(w, "%s\t%d\n", r, data.ByReason[r])
				}
				return w.Flush()
			}

			return nil
		},
	}
}
