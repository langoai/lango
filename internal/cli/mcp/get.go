package mcp

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	mcplib "github.com/langoai/lango/internal/mcp"
)

func newGetCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Show details of an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			merged, err := mcplib.MergedServersStrict(&cfg.MCP)
			if err != nil {
				return err
			}
			srv, ok := merged[name]
			if !ok {
				return fmt.Errorf("server %q not found", name)
			}

			transport := srv.Transport
			if transport == "" {
				transport = "stdio"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Server: %s\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Transport:    %s\n", transport)
			fmt.Fprintf(cmd.OutOrStdout(), "  Enabled:      %v\n", srv.IsEnabled())
			fmt.Fprintf(cmd.OutOrStdout(), "  Safety Level: %s\n", safetylevel(srv.SafetyLevel))

			switch transport {
			case "stdio":
				fmt.Fprintf(cmd.OutOrStdout(), "  Command:      %s\n", srv.Command)
				if len(srv.Args) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "  Args:         %v\n", srv.Args)
				}
				if len(srv.Env) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "  Env vars:     %d configured\n", len(srv.Env))
				}
			case "http", "sse":
				fmt.Fprintf(cmd.OutOrStdout(), "  URL:          %s\n", srv.URL)
				if len(srv.Headers) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "  Headers:      %d configured\n", len(srv.Headers))
				}
			}

			if srv.Timeout > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Timeout:      %s\n", srv.Timeout)
			}

			// Try to connect and list tools
			if !srv.IsEnabled() {
				fmt.Fprintln(cmd.OutOrStdout(), "\n  (server is disabled)")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "\n  Connecting to discover tools...")
			conn := mcplib.NewServerConnection(name, srv, cfg.MCP)
			if err := conn.Connect(context.Background()); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  Connection: FAILED (%v)\n", err)
				return nil
			}
			defer func() { _ = conn.Disconnect(context.Background()) }()

			tools := conn.Tools()
			fmt.Fprintf(cmd.OutOrStdout(), "  Tools:        %d available\n", len(tools))
			for _, dt := range tools {
				desc := dt.Tool.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "    - mcp__%s__%s: %s\n", name, dt.Tool.Name, desc)
			}

			return nil
		},
	}
}

func safetylevel(s string) string {
	if s == "" {
		return "dangerous (default)"
	}
	return s
}
