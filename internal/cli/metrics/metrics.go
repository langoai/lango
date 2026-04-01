// Package metrics provides CLI commands for observability metrics.
package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

const defaultGatewayAddr = "http://localhost:18789"

// NewMetricsCmd creates the metrics command group.
func NewMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "View system observability metrics",
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
		RunE: summaryRunE,
	}

	cmd.PersistentFlags().String("output", "table", "Output format: table or json")
	cmd.PersistentFlags().String("addr", defaultGatewayAddr, "Gateway address")

	cmd.AddCommand(newSessionsCmd())
	cmd.AddCommand(newToolsCmd())
	cmd.AddCommand(newAgentsCmd())
	cmd.AddCommand(newHistoryCmd())
	cmd.AddCommand(newPolicyCmd())

	return cmd
}

func fetchJSON(addr, path string, out interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(addr + path)
	if err != nil {
		return fmt.Errorf("connect to gateway: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gateway returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func getOutputFormat(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("output")
	if f == "" {
		f = "table"
	}
	return f
}

func getAddr(cmd *cobra.Command) string {
	a, _ := cmd.Flags().GetString("addr")
	if a == "" {
		a = defaultGatewayAddr
	}
	return a
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
}

func summaryRunE(cmd *cobra.Command, _ []string) error {
	addr := getAddr(cmd)
	format := getOutputFormat(cmd)

	var snap map[string]interface{}
	if err := fetchJSON(addr, "/metrics", &snap); err != nil {
		return err
	}

	if format == "json" {
		return printJSON(snap)
	}

	fmt.Println("=== System Metrics ===")
	fmt.Println()

	if uptime, ok := snap["uptime"].(string); ok {
		fmt.Printf("Uptime:           %s\n", uptime)
	}
	if tokens, ok := snap["tokenUsage"].(map[string]interface{}); ok {
		fmt.Printf("Total Input:      %.0f tokens\n", toFloat(tokens["inputTokens"]))
		fmt.Printf("Total Output:     %.0f tokens\n", toFloat(tokens["outputTokens"]))
	}
	if execs, ok := snap["toolExecutions"]; ok {
		fmt.Printf("Tool Executions:  %.0f\n", toFloat(execs))
	}

	return nil
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
