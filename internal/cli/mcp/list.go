package mcp

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	mcplib "github.com/langoai/lango/internal/mcp"
)

func newListCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			merged := mcplib.MergedServers(&cfg.MCP)
			if len(merged) == 0 {
				fmt.Println("No MCP servers configured.")
				fmt.Println("\nAdd one with: lango mcp add <name> --type stdio --command <cmd>")
				return nil
			}

			// Sort by name for consistent output
			names := make([]string, 0, len(merged))
			for n := range merged {
				names = append(names, n)
			}
			sort.Strings(names)

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tENABLED\tENDPOINT")
			for _, name := range names {
				srv := merged[name]
				transport := srv.Transport
				if transport == "" {
					transport = "stdio"
				}
				enabled := "yes"
				if !srv.IsEnabled() {
					enabled = "no"
				}
				endpoint := srv.Command
				if srv.URL != "" {
					endpoint = srv.URL
				}
				if len(endpoint) > 60 {
					endpoint = endpoint[:57] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, transport, enabled, endpoint)
			}
			return w.Flush()
		},
	}
}
