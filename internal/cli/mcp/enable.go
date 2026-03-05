package mcp

import (
	"fmt"

	"github.com/spf13/cobra"

	mcplib "github.com/langoai/lango/internal/mcp"
)

func newEnableCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "enable <name>",
		Short: "Enable an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return toggleServer(args[0], scope, true)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "scope: user or project (default: search all)")
	return cmd
}

func newDisableCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "disable <name>",
		Short: "Disable an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return toggleServer(args[0], scope, false)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "scope: user or project (default: search all)")
	return cmd
}

func toggleServer(name, scope string, enabled bool) error {
	paths := scopePaths(scope)
	for _, sp := range paths {
		servers, err := mcplib.LoadMCPFile(sp.path)
		if err != nil {
			continue
		}
		srv, exists := servers[name]
		if !exists {
			continue
		}

		srv.Enabled = boolPtr(enabled)
		servers[name] = srv

		if err := mcplib.SaveMCPFile(sp.path, servers); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		action := "enabled"
		if !enabled {
			action = "disabled"
		}
		fmt.Printf("MCP server %q %s.\n", name, action)
		return nil
	}

	return fmt.Errorf("server %q not found in any scope", name)
}

func boolPtr(b bool) *bool {
	return &b
}
