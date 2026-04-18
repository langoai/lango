package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/turntrace"
)

func traceStoreFromBoot(boot *bootstrap.Result) turntrace.Store {
	if boot != nil && boot.Storage != nil {
		return boot.Storage.TurnTrace()
	}
	return nil
}

func newTraceCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trace",
		Short: "Inspect turn traces",
	}
	cmd.AddCommand(newTraceListCmd(bootLoader))
	cmd.AddCommand(newTraceDetailCmd(bootLoader))
	return cmd
}

func newTraceListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		sessionKey string
		outcomeStr string
		limit      int
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent turn traces",
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

			var traces []turntrace.Trace

			if sessionKey != "" {
				traces, err = store.TracesForSession(ctx, sessionKey)
			} else if outcomeStr != "" {
				since := time.Now().Add(-24 * time.Hour)
				traces, err = store.RecentByOutcome(ctx, turntrace.Outcome(outcomeStr), since, limit)
			} else {
				// Default: recent failures
				traces, err = store.RecentFailures(ctx, limit)
			}
			if err != nil {
				return fmt.Errorf("query traces: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(traces)
			}

			if len(traces) == 0 {
				fmt.Println("No traces found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TRACE ID\tOUTCOME\tSESSION\tDURATION\tSTARTED")
			for _, t := range traces {
				dur := "-"
				if t.EndedAt != nil {
					dur = t.EndedAt.Sub(t.StartedAt).Truncate(time.Millisecond).String()
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					truncStr(t.TraceID, 12),
					t.Outcome,
					truncStr(t.SessionKey, 20),
					dur,
					t.StartedAt.Format("15:04:05"),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&sessionKey, "session", "", "Filter by session key")
	cmd.Flags().StringVar(&outcomeStr, "outcome", "", "Filter by outcome (e.g., timeout, loop_detected)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum traces to return")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newTraceDetailCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <trace-id>",
		Short: "Show detailed event timeline for a trace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			traceID := args[0]

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

			events, err := store.EventsForTrace(ctx, traceID)
			if err != nil {
				return fmt.Errorf("query events: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(events)
			}

			if len(events) == 0 {
				fmt.Printf("No events found for trace %s\n", traceID)
				return nil
			}

			fmt.Printf("Trace: %s (%d events)\n\n", traceID, len(events))
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SEQ\tTIME\tTYPE\tAGENT\tTOOL\tPAYLOAD")
			for _, ev := range events {
				payload := truncStr(ev.PayloadJSON, 60)
				if ev.PayloadTruncated {
					payload += " [truncated]"
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
					ev.Seq,
					ev.CreatedAt.Format("15:04:05.000"),
					ev.EventType,
					ev.AgentName,
					ev.ToolName,
					payload,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func truncStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
