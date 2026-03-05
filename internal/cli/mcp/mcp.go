// Package mcp provides CLI commands for MCP server management.
package mcp

import (
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

// NewMCPCmd creates the mcp command with lazy bootstrap loading.
func NewMCPCmd(
	cfgLoader func() (*config.Config, error),
	bootLoader func() (*bootstrap.Result, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP (Model Context Protocol) servers",
		Long: `Manage external MCP servers that provide tools, resources, and prompts.

MCP servers extend the agent with additional capabilities by connecting to
external processes or HTTP endpoints that implement the Model Context Protocol.

Examples:
  lango mcp list                          # List configured servers
  lango mcp add github --type stdio ...   # Add a server
  lango mcp test github                   # Test server connectivity
  lango mcp get github                    # Show server details`,
	}

	cmd.AddCommand(newListCmd(cfgLoader))
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newRemoveCmd())
	cmd.AddCommand(newGetCmd(cfgLoader))
	cmd.AddCommand(newTestCmd(cfgLoader))
	cmd.AddCommand(newEnableCmd())
	cmd.AddCommand(newDisableCmd())

	return cmd
}
