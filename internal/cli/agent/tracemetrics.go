package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/turntrace"
)

func newTraceMetricsCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		agentFilter string
		jsonOutput  bool
	)

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show trace-derived per-agent performance metrics",
		Long:  "Displays per-agent success rate, turn count, tool calls, and duration percentiles derived from turn traces. Distinct from 'lango metrics agents' which shows token usage.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			store := traceStoreFromBoot(boot)
			if store == nil {
				return fmt.Errorf("trace store unavailable")
			}
			ctx := cmd.Context()

			// Get recent successful and failed traces.
			since := time.Now().Add(-24 * time.Hour)
			successTraces, err := store.RecentByOutcome(ctx, turntrace.OutcomeSuccess, since, 500)
			if err != nil {
				return fmt.Errorf("query success traces: %w", err)
			}
			failureTraces, err := store.RecentFailures(ctx, 500)
			if err != nil {
				return fmt.Errorf("query failure traces: %w", err)
			}

			allTraces := append(successTraces, failureTraces...)

			var allEvents []turntrace.Event
			for _, t := range allTraces {
				events, err := store.EventsForTrace(ctx, t.TraceID)
				if err != nil {
					continue
				}
				allEvents = append(allEvents, events...)
			}

			metrics := turntrace.ComputeAgentMetrics(allTraces, allEvents)

			if agentFilter != "" {
				filtered := make(map[string]*turntrace.AgentMetricsSummary)
				if m, ok := metrics[agentFilter]; ok {
					filtered[agentFilter] = m
				}
				metrics = filtered
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(metrics)
			}

			if len(metrics) == 0 {
				fmt.Println("No agent metrics found.")
				return nil
			}

			// Sort by name for stable output.
			names := make([]string, 0, len(metrics))
			for name := range metrics {
				names = append(names, name)
			}
			sort.Strings(names)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "AGENT\tTURNS\tSUCCESS\tFAILURE\tRATE\tTOOLS\tP50\tP95")
			for _, name := range names {
				m := metrics[name]
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%.0f%%\t%d\t%s\t%s\n",
					name,
					m.TotalTurns,
					m.SuccessCount,
					m.FailureCount,
					m.SuccessRate*100,
					m.ToolCallCount,
					m.P50Duration.Truncate(time.Millisecond),
					m.P95Duration.Truncate(time.Millisecond),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&agentFilter, "agent", "", "Filter by agent name")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
