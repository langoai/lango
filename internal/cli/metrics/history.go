package metrics

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/clihttp"
)

func newHistoryCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:           "history",
		Short:         "Historical token usage from database",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format, err := getOutputFormat(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/metrics/history?days=%d", days)

			var data struct {
				Records []struct {
					Provider     string    `json:"provider"`
					Model        string    `json:"model"`
					SessionKey   string    `json:"sessionKey"`
					AgentName    string    `json:"agentName"`
					InputTokens  int64     `json:"inputTokens"`
					OutputTokens int64     `json:"outputTokens"`
					Timestamp    time.Time `json:"timestamp"`
				} `json:"records"`
				Total struct {
					InputTokens  int64 `json:"inputTokens"`
					OutputTokens int64 `json:"outputTokens"`
					RecordCount  int   `json:"recordCount"`
				} `json:"total"`
			}
			if err := clihttp.FetchJSON(addr, path, &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(cmd.OutOrStdout(), data)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Token usage history (last %d days)\n", days)
			fmt.Fprintf(cmd.OutOrStdout(), "Records: %d | Total Input: %d | Total Output: %d\n\n",
				data.Total.RecordCount, data.Total.InputTokens, data.Total.OutputTokens)

			if len(data.Records) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No historical data available.")
				return nil
			}

			w := newTabWriter(cmd.OutOrStdout())
			fmt.Fprintln(w, "TIME\tPROVIDER\tMODEL\tINPUT\tOUTPUT")
			for _, r := range data.Records {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n",
					r.Timestamp.Format("2006-01-02 15:04"),
					r.Provider, truncate(r.Model, 20),
					r.InputTokens, r.OutputTokens)
			}
			return w.Flush()
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "Number of days to query")
	return cmd
}
