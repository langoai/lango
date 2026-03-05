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

			merged := mcplib.MergedServers(&cfg.MCP)
			srv, ok := merged[name]
			if !ok {
				return fmt.Errorf("server %q not found", name)
			}

			transport := srv.Transport
			if transport == "" {
				transport = "stdio"
			}

			fmt.Printf("Testing %q...\n", name)
			fmt.Printf("  Transport:  %s", transport)
			if transport == "stdio" {
				fmt.Printf(" (%s", srv.Command)
				for _, a := range srv.Args {
					fmt.Printf(" %s", a)
				}
				fmt.Print(")")
			} else {
				fmt.Printf(" (%s)", srv.URL)
			}
			fmt.Println()

			// Test connection
			conn := mcplib.NewServerConnection(name, srv, cfg.MCP)

			start := time.Now()
			if err := conn.Connect(context.Background()); err != nil {
				fmt.Printf("  Handshake:  FAILED (%v)\n", err)
				return nil
			}
			handshake := time.Since(start)
			fmt.Printf("  Handshake:  OK (%s)\n", handshake.Truncate(time.Millisecond))

			defer func() { _ = conn.Disconnect(context.Background()) }()

			// List tools
			tools := conn.Tools()
			fmt.Printf("  Tools:      %d available\n", len(tools))

			// Ping
			session := conn.Session()
			if session != nil {
				pingStart := time.Now()
				if err := session.Ping(context.Background(), nil); err != nil {
					fmt.Printf("  Ping:       FAILED (%v)\n", err)
				} else {
					fmt.Printf("  Ping:       OK (%s)\n", time.Since(pingStart).Truncate(time.Millisecond))
				}
			}

			return nil
		},
	}
}
