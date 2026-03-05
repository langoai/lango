package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	mcplib "github.com/langoai/lango/internal/mcp"
)

func newAddCmd() *cobra.Command {
	var (
		transport string
		command   string
		rawArgs   string
		url       string
		env       []string
		headers   []string
		scope     string
		safety    string
	)

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new MCP server",
		Long: `Add a new MCP server configuration.

Examples:
  # Add stdio server
  lango mcp add github --type stdio \
    --command npx --args "-y,@modelcontextprotocol/server-github" \
    --env "GITHUB_TOKEN=\${GITHUB_TOKEN}" \
    --scope project

  # Add HTTP server
  lango mcp add remote-api --type http \
    --url "https://api.example.com/mcp" \
    --header "Authorization=Bearer \${TOKEN}" \
    --scope user`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if transport == "" {
				transport = "stdio"
			}

			srv := config.MCPServerConfig{
				Transport:   transport,
				Command:     command,
				URL:         url,
				SafetyLevel: safety,
			}

			if rawArgs != "" {
				srv.Args = strings.Split(rawArgs, ",")
			}

			if len(env) > 0 {
				srv.Env = parseKV(env)
			}
			if len(headers) > 0 {
				srv.Headers = parseKV(headers)
			}

			// Validate
			switch transport {
			case "stdio":
				if command == "" {
					return fmt.Errorf("--command is required for stdio transport")
				}
			case "http", "sse":
				if url == "" {
					return fmt.Errorf("--url is required for %s transport", transport)
				}
			default:
				return fmt.Errorf("invalid transport type: %s (must be stdio, http, or sse)", transport)
			}

			// Determine file path based on scope
			path, err := scopePath(scope)
			if err != nil {
				return err
			}

			// Load existing, add, save
			servers, _ := mcplib.LoadMCPFile(path)
			if servers == nil {
				servers = make(map[string]config.MCPServerConfig)
			}
			if _, exists := servers[name]; exists {
				return fmt.Errorf("server %q already exists in %s scope", name, scope)
			}
			servers[name] = srv

			if err := mcplib.SaveMCPFile(path, servers); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("MCP server %q added to %s scope (%s).\n", name, scope, path)
			fmt.Printf("  Transport: %s\n", transport)
			if command != "" {
				fmt.Printf("  Command:   %s %s\n", command, rawArgs)
			}
			if url != "" {
				fmt.Printf("  URL:       %s\n", url)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&transport, "type", "stdio", "transport type: stdio, http, sse")
	cmd.Flags().StringVar(&command, "command", "", "executable command (stdio)")
	cmd.Flags().StringVar(&rawArgs, "args", "", "comma-separated arguments (stdio)")
	cmd.Flags().StringVar(&url, "url", "", "endpoint URL (http/sse)")
	cmd.Flags().StringSliceVar(&env, "env", nil, "environment variables (KEY=VALUE)")
	cmd.Flags().StringSliceVar(&headers, "header", nil, "HTTP headers (KEY=VALUE)")
	cmd.Flags().StringVar(&scope, "scope", "user", "config scope: user or project")
	cmd.Flags().StringVar(&safety, "safety", "dangerous", "safety level: safe, moderate, dangerous")

	return cmd
}

func parseKV(pairs []string) map[string]string {
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if ok {
			m[k] = v
		}
	}
	return m
}

func scopePath(scope string) (string, error) {
	switch scope {
	case "project":
		return ".lango-mcp.json", nil
	case "user", "":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		dir := filepath.Join(home, ".lango")
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", fmt.Errorf("create config directory: %w", err)
		}
		return filepath.Join(dir, "mcp.json"), nil
	default:
		return "", fmt.Errorf("invalid scope: %s (must be user or project)", scope)
	}
}
