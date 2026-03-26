package mcp

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

// ServerManager manages multiple MCP server connections.
type ServerManager struct {
	cfg      config.MCPConfig
	mu       sync.RWMutex
	servers  map[string]*ServerConnection
	isolator sandboxos.OSIsolator // OS-level sandbox for stdio server processes (nil = disabled)
}

// SetOSIsolator configures OS-level sandbox for all managed stdio server connections.
func (m *ServerManager) SetOSIsolator(iso sandboxos.OSIsolator) {
	m.isolator = iso
}

// NewServerManager creates a new manager for the given config.
func NewServerManager(cfg config.MCPConfig) *ServerManager {
	return &ServerManager{
		cfg:     cfg,
		servers: make(map[string]*ServerConnection),
	}
}

// ConnectAll connects to all configured and enabled servers.
// Returns a map of server names to errors for any that failed.
func (m *ServerManager) ConnectAll(ctx context.Context) map[string]error {
	errs := make(map[string]error)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, srvCfg := range m.cfg.Servers {
		if !srvCfg.IsEnabled() {
			logging.App().Infow("MCP server disabled, skipping", "server", name)
			continue
		}

		conn := NewServerConnection(name, srvCfg, m.cfg)
		if m.isolator != nil {
			conn.SetOSIsolator(m.isolator)
		}
		m.mu.Lock()
		m.servers[name] = conn
		m.mu.Unlock()

		wg.Add(1)
		go func(n string, c *ServerConnection) {
			defer wg.Done()
			if err := c.Connect(ctx); err != nil {
				mu.Lock()
				errs[n] = err
				mu.Unlock()
				logging.App().Warnw("MCP server connection failed", "server", n, "error", err)
			} else {
				c.StartHealthCheck(ctx)
			}
		}(name, conn)
	}

	wg.Wait()
	return errs
}

// DisconnectAll disconnects all managed servers.
func (m *ServerManager) DisconnectAll(ctx context.Context) error {
	m.mu.RLock()
	servers := make([]*ServerConnection, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	m.mu.RUnlock()

	for _, s := range servers {
		if err := s.Disconnect(ctx); err != nil {
			logging.App().Warnw("MCP server disconnect error", "server", s.Name(), "error", err)
		}
	}
	return nil
}

// AllTools returns all discovered tools from all connected servers.
func (m *ServerManager) AllTools() []DiscoveredTool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []DiscoveredTool
	for _, s := range m.servers {
		all = append(all, s.Tools()...)
	}
	return all
}

// AllResources returns all discovered resources from all connected servers.
func (m *ServerManager) AllResources() []DiscoveredResource {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []DiscoveredResource
	for _, s := range m.servers {
		sc := s
		sc.mu.RLock()
		res := make([]DiscoveredResource, len(sc.resources))
		copy(res, sc.resources)
		sc.mu.RUnlock()
		all = append(all, res...)
	}
	return all
}

// AllPrompts returns all discovered prompts from all connected servers.
func (m *ServerManager) AllPrompts() []DiscoveredPrompt {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []DiscoveredPrompt
	for _, s := range m.servers {
		sc := s
		sc.mu.RLock()
		pr := make([]DiscoveredPrompt, len(sc.prompts))
		copy(pr, sc.prompts)
		sc.mu.RUnlock()
		all = append(all, pr...)
	}
	return all
}

// ServerStatus returns the state of each managed server.
func (m *ServerManager) ServerStatus() map[string]ServerState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]ServerState, len(m.servers))
	for name, s := range m.servers {
		status[name] = s.State()
	}
	return status
}

// GetConnection returns the named server connection.
func (m *ServerManager) GetConnection(name string) (*ServerConnection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.servers[name]
	return s, ok
}

// ServerCount returns the number of managed servers.
func (m *ServerManager) ServerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.servers)
}
