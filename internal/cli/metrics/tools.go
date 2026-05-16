package metrics

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/clihttp"
)

func newToolsCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "tools",
		Short:         "Tool execution statistics",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format, err := getOutputFormat(cmd)
			if err != nil {
				return err
			}

			var data struct {
				Tools []struct {
					Name        string  `json:"name"`
					Count       int64   `json:"count"`
					Errors      int64   `json:"errors"`
					AvgDuration string  `json:"avgDuration"`
					ErrorRate   float64 `json:"errorRate"`
				} `json:"tools"`
			}
			if err := clihttp.FetchJSON(addr, "/metrics/tools", &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(cmd.OutOrStdout(), data)
			}

			if len(data.Tools) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No tool execution data available.")
				return nil
			}

			w := newTabWriter(cmd.OutOrStdout())
			fmt.Fprintln(w, "TOOL\tCOUNT\tERRORS\tERROR RATE\tAVG DURATION")
			for _, t := range data.Tools {
				fmt.Fprintf(w, "%s\t%d\t%d\t%.1f%%\t%s\n",
					t.Name, t.Count, t.Errors, t.ErrorRate*100, t.AvgDuration)
			}
			return w.Flush()
		},
	}
}
