package metrics

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/clihttp"
)

func newSessionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "sessions",
		Short:         "Per-session token usage breakdown",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format, err := getOutputFormat(cmd)
			if err != nil {
				return err
			}

			var data struct {
				Sessions []struct {
					SessionKey   string `json:"sessionKey"`
					InputTokens  int64  `json:"inputTokens"`
					OutputTokens int64  `json:"outputTokens"`
					TotalTokens  int64  `json:"totalTokens"`
					RequestCount int64  `json:"requestCount"`
				} `json:"sessions"`
			}
			if err := clihttp.FetchJSON(addr, "/metrics/sessions", &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(cmd.OutOrStdout(), data)
			}

			if len(data.Sessions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No session data available.")
				return nil
			}

			w := newTabWriter(cmd.OutOrStdout())
			fmt.Fprintln(w, "SESSION\tINPUT\tOUTPUT\tTOTAL\tREQUESTS")
			for _, s := range data.Sessions {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\n",
					truncate(s.SessionKey, 24), s.InputTokens, s.OutputTokens,
					s.TotalTokens, s.RequestCount)
			}
			return w.Flush()
		},
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
