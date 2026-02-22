package checks

import (
	"context"
	"fmt"
	"net"

	"github.com/langoai/lango/internal/config"
)

// NetworkCheck validates network-related configuration.
type NetworkCheck struct{}

// Name returns the check name.
func (c *NetworkCheck) Name() string {
	return "Server Port"
}

// Run checks if the configured server port is available.
func (c *NetworkCheck) Run(ctx context.Context, cfg *config.Config) Result {
	port := 18789 // default
	if cfg != nil && cfg.Server.Port > 0 {
		port = cfg.Server.Port
	}

	host := "localhost"
	if cfg != nil && cfg.Server.Host != "" {
		host = cfg.Server.Host
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	// Try to listen on the port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("Port %d is not available", port),
			Details: err.Error(),
		}
	}
	listener.Close()

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("Port %d available", port),
		Details: addr,
	}
}

// Fix cannot auto-fix port conflicts.
func (c *NetworkCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return Result{
		Name:    c.Name(),
		Status:  StatusSkip,
		Message: "Port conflicts require manual resolution",
		Details: "Change server.port in config or stop the conflicting process",
	}
}
