package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/turntrace"
)

func newGraphCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "graph <session-key>",
		Short: "Show delegation graph for a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionKey := args[0]

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			store := turntrace.NewEntStore(boot.DBClient)
			ctx := cmd.Context()

			traces, err := store.TracesForSession(ctx, sessionKey)
			if err != nil {
				return fmt.Errorf("query traces: %w", err)
			}

			var allEvents []turntrace.Event
			for _, t := range traces {
				events, err := store.EventsForTrace(ctx, t.TraceID)
				if err != nil {
					return fmt.Errorf("query events for trace %q: %w", t.TraceID, err)
				}
				allEvents = append(allEvents, events...)
			}

			graph := turntrace.BuildDelegationGraph(sessionKey, traces, allEvents)

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(graph)
			}

			if len(graph.Agents) == 0 {
				fmt.Println("No delegation data found for this session.")
				return nil
			}

			fmt.Printf("Delegation graph for session: %s\n\n", sessionKey)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "AGENT\tIN\tOUT\tTOOL CALLS")
			for _, node := range graph.Agents {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\n",
					node.Name, node.DelegationsIn, node.DelegationsOut, node.ToolCalls)
			}
			_ = w.Flush()

			if len(graph.Edges) > 0 {
				fmt.Printf("\nEdges (%d):\n", len(graph.Edges))
				for _, edge := range graph.Edges {
					fmt.Printf("  %s → %s  (%s)\n", edge.From, edge.To, edge.Timestamp.Format("15:04:05"))
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
