package metrics

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSessionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sessions",
		Short: "Per-session token usage breakdown",
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format := getOutputFormat(cmd)

			var data struct {
				Sessions []struct {
					SessionKey   string `json:"sessionKey"`
					InputTokens  int64  `json:"inputTokens"`
					OutputTokens int64  `json:"outputTokens"`
					TotalTokens  int64  `json:"totalTokens"`
					RequestCount int64  `json:"requestCount"`
				} `json:"sessions"`
			}
			if err := fetchJSON(addr, "/metrics/sessions", &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(data)
			}

			if len(data.Sessions) == 0 {
				fmt.Println("No session data available.")
				return nil
			}

			w := newTabWriter()
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
