package metrics

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAgentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "agents",
		Short: "Per-agent token usage breakdown",
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format := getOutputFormat(cmd)

			var data struct {
				Agents []struct {
					Name         string `json:"name"`
					InputTokens  int64  `json:"inputTokens"`
					OutputTokens int64  `json:"outputTokens"`
					ToolCalls    int64  `json:"toolCalls"`
				} `json:"agents"`
			}
			if err := fetchJSON(addr, "/metrics/agents", &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(data)
			}

			if len(data.Agents) == 0 {
				fmt.Println("No agent data available.")
				return nil
			}

			w := newTabWriter()
			fmt.Fprintln(w, "AGENT\tINPUT\tOUTPUT\tTOOL CALLS")
			for _, a := range data.Agents {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\n",
					a.Name, a.InputTokens, a.OutputTokens, a.ToolCalls)
			}
			return w.Flush()
		},
	}
}
