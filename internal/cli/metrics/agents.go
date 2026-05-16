package metrics

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/clihttp"
)

func newAgentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "agents",
		Short:         "Per-agent token usage breakdown",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format, err := getOutputFormat(cmd)
			if err != nil {
				return err
			}

			var data struct {
				Agents []struct {
					Name         string `json:"name"`
					InputTokens  int64  `json:"inputTokens"`
					OutputTokens int64  `json:"outputTokens"`
					ToolCalls    int64  `json:"toolCalls"`
				} `json:"agents"`
			}
			if err := clihttp.FetchJSON(addr, "/metrics/agents", &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(cmd.OutOrStdout(), data)
			}

			if len(data.Agents) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No agent data available.")
				return nil
			}

			w := newTabWriter(cmd.OutOrStdout())
			fmt.Fprintln(w, "AGENT\tINPUT\tOUTPUT\tTOOL CALLS")
			for _, a := range data.Agents {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\n",
					a.Name, a.InputTokens, a.OutputTokens, a.ToolCalls)
			}
			return w.Flush()
		},
	}
}
