package mcp

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/logging"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

// ServerManager manages multiple MCP server connections.
type ServerManager struct {
	cfg           config.MCPConfig
	mu            sync.RWMutex
	servers       map[string]*ServerConnection
	isolator      sandboxos.OSIsolator
	failClosed    bool
	workspacePath string        // User workspace root, forwarded to each connection for MCPServerPolicy walk-up
	dataRoot      string        // Lango control-plane root, forwarded to each connection
	bus           *eventbus.Bus // event bus, forwarded to each connection
}

// NewServerManager creates a new manager for the given config.
func NewServerManager(cfg config.MCPConfig) *ServerManager {
	return &ServerManager{
		cfg:     cfg,
		servers: make(map[string]*ServerConnection),
	}
}

// SetOSIsolator sets the OS-level sandbox isolator for all current
// and future connections. workspacePath is forwarded so each connection's
// MCPServerPolicy can walk up to the repo `.git` and apply the same
// baseline deny as DefaultToolPolicy. dataRoot is forwarded so each
// connection's policy denies the lango control-plane to the spawned MCP
// child.
func (m *ServerManager) SetOSIsolator(iso sandboxos.OSIsolator, workspacePath, dataRoot string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isolator = iso
	m.workspacePath = workspacePath
	m.dataRoot = dataRoot
	for _, s := range m.servers {
		s.SetOSIsolator(iso, workspacePath, dataRoot)
	}
}

// SetFailClosed enables or disables fail-closed semantics for all current
// and future connections. When true, stdio MCP transport creation is blocked
// if no OS sandbox can be applied.
func (m *ServerManager) SetFailClosed(fc bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failClosed = fc
	for _, s := range m.servers {
		s.SetFailClosed(fc)
	}
}

// SetEventBus attaches an event bus for SandboxDecisionEvent publishing on
// all current and future connections.
func (m *ServerManager) SetEventBus(bus *eventbus.Bus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bus = bus
	for _, s := range m.servers {
		s.SetEventBus(bus)
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
			conn.SetOSIsolator(m.isolator, m.workspacePath, m.dataRoot)
		}
		if m.bus != nil {
			conn.SetEventBus(m.bus)
		}
		conn.SetFailClosed(m.failClosed)
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
