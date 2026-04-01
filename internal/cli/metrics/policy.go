package metrics

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newPolicyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "policy",
		Short: "Policy decision statistics",
		Long: `View policy decision metrics including block and observe counts
with per-reason breakdowns.

Examples:
  lango metrics policy                  # Table summary
  lango metrics policy --output json    # JSON output`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format := getOutputFormat(cmd)

			var data struct {
				Blocks   int64            `json:"blocks"`
				Observes int64            `json:"observes"`
				ByReason map[string]int64 `json:"byReason"`
			}
			if err := fetchJSON(addr, "/metrics/policy", &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(data)
			}

			fmt.Println("=== Policy Decisions ===")
			fmt.Println()
			fmt.Printf("Blocks:    %d\n", data.Blocks)
			fmt.Printf("Observes:  %d\n", data.Observes)

			if len(data.ByReason) > 0 {
				fmt.Println()
				fmt.Println("--- By Reason ---")
				w := newTabWriter()
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
