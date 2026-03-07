package metrics

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newToolsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tools",
		Short: "Tool execution statistics",
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr := getAddr(cmd)
			format := getOutputFormat(cmd)

			var data struct {
				Tools []struct {
					Name        string  `json:"name"`
					Count       int64   `json:"count"`
					Errors      int64   `json:"errors"`
					AvgDuration string  `json:"avgDuration"`
					ErrorRate   float64 `json:"errorRate"`
				} `json:"tools"`
			}
			if err := fetchJSON(addr, "/metrics/tools", &data); err != nil {
				return err
			}

			if format == "json" {
				return printJSON(data)
			}

			if len(data.Tools) == 0 {
				fmt.Println("No tool execution data available.")
				return nil
			}

			w := newTabWriter()
			fmt.Fprintln(w, "TOOL\tCOUNT\tERRORS\tERROR RATE\tAVG DURATION")
			for _, t := range data.Tools {
				fmt.Fprintf(w, "%s\t%d\t%d\t%.1f%%\t%s\n",
					t.Name, t.Count, t.Errors, t.ErrorRate*100, t.AvgDuration)
			}
			return w.Flush()
		},
	}
}
