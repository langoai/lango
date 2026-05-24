// Package metrics provides CLI commands for observability metrics.
package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/clihttp"
	"github.com/langoai/lango/internal/config"
)


const defaultGatewayAddr = clihttp.DefaultGatewayAddr

type configLoader func() (*config.Config, error)

// NewMetricsCmd creates the metrics command group.
func NewMetricsCmd() *cobra.Command {
	return NewMetricsCmdWithConfig(nil)
}

// NewMetricsCmdWithConfig creates the metrics command group with config-backed gateway defaults.
func NewMetricsCmdWithConfig(loadConfig configLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "metrics",
		Short:         "View system observability metrics",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `View system observability metrics including token usage, tool execution stats,
and agent performance.

Requires a running Lango server (lango serve).

Examples:
  lango metrics                        # System snapshot summary
  lango metrics sessions               # Per-session token breakdown
  lango metrics tools                  # Tool execution statistics
  lango metrics agents                 # Per-agent token usage
  lango metrics history --days=7       # Historical token usage
  lango metrics policy                 # Policy decision statistics`,
		RunE: summaryRunEWithConfig(loadConfig),
	}

	cmd.PersistentFlags().String("output", "table", "Output format: table or json")
	cmd.PersistentFlags().String("addr", "", "Gateway address (default: configured server host/port)")

	cmd.AddCommand(newSessionsCmd(loadConfig))
	cmd.AddCommand(newToolsCmd(loadConfig))
	cmd.AddCommand(newAgentsCmd(loadConfig))
	cmd.AddCommand(newHistoryCmd(loadConfig))
	cmd.AddCommand(newPolicyCmd(loadConfig))

	return cmd
}

func getOutputFormat(cmd *cobra.Command) (string, error) {
	return clihttp.ResolveTableOrJSONOutput(cmd)
}

func getAddr(cmd *cobra.Command, loadConfig configLoader) (string, error) {
	a, _ := cmd.Flags().GetString("addr")
	if a != "" || loadConfig == nil {
		return clihttp.ResolveGatewayAddr(a, nil), nil
	}
	cfg, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}
	return clihttp.ResolveGatewayAddr("", cfg), nil
}

func printJSON(w io.Writer, v interface{}) error {
	return clihttp.PrintJSON(w, v)
}

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
}

func summaryRunEWithConfig(loadConfig configLoader) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		format, err := getOutputFormat(cmd)
		if err != nil {
			return err
		}
		addr, err := getAddr(cmd, loadConfig)
		if err != nil {
			return err
		}

		var snap map[string]interface{}
		if err := clihttp.FetchJSON(addr, "/metrics", &snap); err != nil {
			return err
		}

		if format == "json" {
			return printJSON(cmd.OutOrStdout(), snap)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "=== System Metrics ===")
		fmt.Fprintln(cmd.OutOrStdout())

		if uptime, ok := snap["uptime"].(string); ok {
			fmt.Fprintf(cmd.OutOrStdout(), "Uptime:           %s\n", uptime)
		}
		if tokens, ok := snap["tokenUsage"].(map[string]interface{}); ok {
			fmt.Fprintf(cmd.OutOrStdout(), "Total Input:      %.0f tokens\n", toFloat(tokens["inputTokens"]))
			fmt.Fprintf(cmd.OutOrStdout(), "Total Output:     %.0f tokens\n", toFloat(tokens["outputTokens"]))
		}
		if execs, ok := snap["toolExecutions"]; ok {
			fmt.Fprintf(cmd.OutOrStdout(), "Tool Executions:  %.0f\n", toFloat(execs))
		}

		return nil
	}
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}
