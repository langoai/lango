package mcp

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	mcplib "github.com/langoai/lango/internal/mcp"
)

func newRemoveCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove an MCP server configuration",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Try to find and remove from the specified scope
			paths := scopePaths(scope)
			for _, sp := range paths {
				servers, err := mcplib.LoadMCPFile(sp.path)
				if err != nil {
					continue
				}
				if _, exists := servers[name]; !exists {
					continue
				}

				delete(servers, name)
				if err := mcplib.SaveMCPFile(sp.path, servers); err != nil {
					return fmt.Errorf("save config: %w", err)
				}
				fmt.Printf("MCP server %q removed from %s scope.\n", name, sp.scope)
				return nil
			}

			return fmt.Errorf("server %q not found in any scope", name)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "scope to remove from: user or project (default: search all)")
	return cmd
}

type scopeInfo struct {
	scope string
	path  string
}

func scopePaths(scope string) []scopeInfo {
	var paths []scopeInfo

	if scope == "" || scope == "project" {
		paths = append(paths, scopeInfo{scope: "project", path: ".lango-mcp.json"})
	}
	if scope == "" || scope == "user" {
		if home, err := os.UserHomeDir(); err == nil {
			paths = append(paths, scopeInfo{scope: "user", path: filepath.Join(home, ".lango", "mcp.json")})
		}
	}

	return paths
}
