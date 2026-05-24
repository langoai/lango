package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	mcplib "github.com/langoai/lango/internal/mcp"
)

func newTestCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "test <name>",
		Short: "Test connectivity to an MCP server",
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

			fmt.Fprintf(cmd.OutOrStdout(), "Testing %q...\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Transport:  %s", transport)
			if transport == "stdio" {
				fmt.Fprintf(cmd.OutOrStdout(), " (%s", srv.Command)
				for _, a := range srv.Args {
					fmt.Fprintf(cmd.OutOrStdout(), " %s", a)
				}
				fmt.Fprint(cmd.OutOrStdout(), ")")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), " (%s)", srv.URL)
			}
			fmt.Fprintln(cmd.OutOrStdout())

			// Test connection
			conn := mcplib.NewServerConnection(name, srv, cfg.MCP)

			start := time.Now()
			if err := conn.Connect(context.Background()); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  Handshake:  FAILED (%v)\n", err)
				return nil
			}
			handshake := time.Since(start)
			fmt.Fprintf(cmd.OutOrStdout(), "  Handshake:  OK (%s)\n", handshake.Truncate(time.Millisecond))

			defer func() { _ = conn.Disconnect(context.Background()) }()

			// List tools
			tools := conn.Tools()
			fmt.Fprintf(cmd.OutOrStdout(), "  Tools:      %d available\n", len(tools))

			// Ping
			session := conn.Session()
			if session != nil {
				pingStart := time.Now()
				if err := session.Ping(context.Background(), nil); err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  Ping:       FAILED (%v)\n", err)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  Ping:       OK (%s)\n", time.Since(pingStart).Truncate(time.Millisecond))
				}
			}

			return nil
		},
	}
}
