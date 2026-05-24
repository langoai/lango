package mcp

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os/exec"
	"sync"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/logging"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

// ServerState represents the lifecycle state of an MCP server connection.
type ServerState int

const (
	StateDisconnected ServerState = iota
	StateConnecting
	StateConnected
	StateFailed
	StateStopped
)

// String returns a human-readable state name.
func (s ServerState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateFailed:
		return "failed"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// DiscoveredTool holds an MCP tool definition along with its source server.
type DiscoveredTool struct {
	ServerName string
	Tool       *sdkmcp.Tool
}

// DiscoveredResource holds an MCP resource definition along with its source server.
type DiscoveredResource struct {
	ServerName string
	Resource   *sdkmcp.Resource
}

// DiscoveredPrompt holds an MCP prompt definition along with its source server.
type DiscoveredPrompt struct {
	ServerName string
	Prompt     *sdkmcp.Prompt
}

// ServerConnection manages the lifecycle of a single MCP server.
type ServerConnection struct {
	name   string
	cfg    config.MCPServerConfig
	global config.MCPConfig

	mu      sync.RWMutex
	state   ServerState
	client  *sdkmcp.Client
	session *sdkmcp.ClientSession

	tools     []DiscoveredTool
	resources []DiscoveredResource
	prompts   []DiscoveredPrompt

	isolator       sandboxos.OSIsolator
	failClosed     bool
	workspacePath  string // User workspace root, used by MCPServerPolicy to walk up to .git
	dataRoot       string // Lango control-plane root, masked from sandboxed MCP child
	protectedPaths []string
	bus            *eventbus.Bus // event bus for SandboxDecisionEvent (optional)

	stopCh chan struct{}
}

// NewServerConnection creates a new server connection manager.
func NewServerConnection(name string, cfg config.MCPServerConfig, global config.MCPConfig) *ServerConnection {
	return &ServerConnection{
		name:   name,
		cfg:    cfg,
		global: global,
		state:  StateDisconnected,
		stopCh: make(chan struct{}),
	}
}

// Name returns the server name.
func (sc *ServerConnection) Name() string { return sc.name }

// SetOSIsolator sets the OS-level sandbox isolator for this connection.
// workspacePath is recorded so MCPServerPolicy can walk up to find the
// repository `.git` and apply the same baseline deny as DefaultToolPolicy.
// dataRoot is recorded so the policy applied at transport creation time can
// deny the lango control-plane to the spawned MCP server child process.
// workspacePath also becomes the MCP child's cmd.Dir so policy discovery
// and execution share the same git context.
func (sc *ServerConnection) SetOSIsolator(iso sandboxos.OSIsolator, workspacePath, dataRoot string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.isolator = iso
	sc.workspacePath = workspacePath
	sc.dataRoot = dataRoot
}

// SetProtectedPaths stores the resolved runtime denylist for policy construction.
func (sc *ServerConnection) SetProtectedPaths(paths []string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.protectedPaths = append([]string(nil), paths...)
}

// SetFailClosed sets whether this connection blocks stdio transport
// creation when no sandbox is available.
func (sc *ServerConnection) SetFailClosed(fc bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.failClosed = fc
}

// SetEventBus attaches an event bus for SandboxDecisionEvent publishing.
func (sc *ServerConnection) SetEventBus(bus *eventbus.Bus) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.bus = bus
}

// publishSandboxDecision publishes a SandboxDecisionEvent for this MCP server's
// transport-creation moment. SessionKey is intentionally empty: MCP server
// lifecycle is process-level, not session-bound.
func (sc *ServerConnection) publishSandboxDecision(decision, reason string) {
	backend := ""
	if sc.isolator != nil {
		backend = sc.isolator.Name()
	}
	eventbus.PublishSandboxDecision(sc.bus, eventbus.SandboxDecisionEvent{
		Source:   "mcp",
		Command:  sc.name,
		Decision: decision,
		Backend:  backend,
		Reason:   reason,
	})
}

// State returns the current connection state.
func (sc *ServerConnection) State() ServerState {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.state
}

// Session returns the active client session, or nil if not connected.
func (sc *ServerConnection) Session() *sdkmcp.ClientSession {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.session
}

// Tools returns the discovered tools from this server.
func (sc *ServerConnection) Tools() []DiscoveredTool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	out := make([]DiscoveredTool, len(sc.tools))
	copy(out, sc.tools)
	return out
}

// Connect establishes a connection to the MCP server and discovers capabilities.
func (sc *ServerConnection) Connect(ctx context.Context) error {
	sc.mu.Lock()
	sc.state = StateConnecting
	sc.mu.Unlock()

	transport, err := sc.createTransport()
	if err != nil {
		sc.setState(StateFailed)
		return fmt.Errorf("%w: %s: %v", ErrConnectionFailed, sc.name, err)
	}

	client := sdkmcp.NewClient(
		&sdkmcp.Implementation{Name: "lango", Version: "1.0.0"},
		nil,
	)

	timeout := sc.timeout()
	connectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	session, err := client.Connect(connectCtx, transport, nil)
	if err != nil {
		sc.setState(StateFailed)
		return fmt.Errorf("%w: %s: %v", ErrConnectionFailed, sc.name, err)
	}

	sc.mu.Lock()
	sc.client = client
	sc.session = session
	sc.state = StateConnected
	sc.mu.Unlock()

	// Discover capabilities
	sc.discoverCapabilities(ctx)

	log := logging.App()
	log.Infow("MCP server connected",
		"server", sc.name,
		"tools", len(sc.tools),
		"resources", len(sc.resources),
		"prompts", len(sc.prompts),
	)

	return nil
}

// Disconnect closes the connection to the MCP server.
func (sc *ServerConnection) Disconnect(ctx context.Context) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Signal health check goroutine to stop
	select {
	case <-sc.stopCh:
	default:
		close(sc.stopCh)
	}

	sc.state = StateStopped

	if sc.session != nil {
		err := sc.session.Close()
		sc.session = nil
		sc.client = nil
		return err
	}
	return nil
}

// StartHealthCheck starts a background goroutine that periodically pings the server.
func (sc *ServerConnection) StartHealthCheck(ctx context.Context) {
	interval := sc.global.HealthCheckInterval
	if interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-sc.stopCh:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				sc.healthCheck(ctx)
			}
		}
	}()
}

func (sc *ServerConnection) healthCheck(ctx context.Context) {
	session := sc.Session()
	if session == nil {
		return
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := session.Ping(pingCtx, nil); err != nil {
		log := logging.App()
		log.Warnw("MCP server health check failed", "server", sc.name, "error", err)
		sc.setState(StateFailed)

		if sc.global.AutoReconnect {
			go sc.reconnect(ctx)
		}
	}
}

func (sc *ServerConnection) reconnect(ctx context.Context) {
	maxAttempts := sc.global.MaxReconnectAttempts
	if maxAttempts <= 0 {
		maxAttempts = 5
	}

	log := logging.App()
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-sc.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		backoff := time.Duration(math.Min(float64(time.Second)*math.Pow(2, float64(attempt-1)), float64(30*time.Second)))
		log.Infow("MCP server reconnecting", "server", sc.name, "attempt", attempt, "backoff", backoff)

		select {
		case <-time.After(backoff):
		case <-sc.stopCh:
			return
		case <-ctx.Done():
			return
		}

		if err := sc.Connect(ctx); err == nil {
			log.Infow("MCP server reconnected", "server", sc.name)
			return
		}
	}

	log.Errorw("MCP server reconnection exhausted", "server", sc.name, "attempts", maxAttempts)
}

func (sc *ServerConnection) setState(s ServerState) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.state = s
}

func (sc *ServerConnection) timeout() time.Duration {
	if sc.cfg.Timeout > 0 {
		return sc.cfg.Timeout
	}
	if sc.global.DefaultTimeout > 0 {
		return sc.global.DefaultTimeout
	}
	return 30 * time.Second
}

func (sc *ServerConnection) createTransport() (sdkmcp.Transport, error) {
	switch sc.cfg.Transport {
	case "", "stdio":
		if sc.cfg.Command == "" {
			return nil, fmt.Errorf("%w: stdio requires command", ErrInvalidTransport)
		}
		cmd := exec.Command(sc.cfg.Command, sc.cfg.Args...)
		if len(sc.cfg.Env) > 0 {
			cmd.Env = BuildEnvSlice(sc.cfg.Env)
		}
		// Set cmd.Dir explicitly so the MCP child runs inside the user's
		// workspace (and so the sandbox policy's git walk-up has a
		// meaningful starting point). An empty workspacePath falls back
		// to the supervisor's cwd — preserves legacy behavior for callers
		// that have not yet migrated to the workspace-aware setter.
		if sc.workspacePath != "" {
			cmd.Dir = sc.workspacePath
		}

		if sc.isolator == nil && sc.failClosed {
			sc.publishSandboxDecision("rejected", "no isolator configured")
			return nil, fmt.Errorf("%w: no OS isolator configured for MCP server %q", sandboxos.ErrSandboxRequired, sc.name)
		}
		if sc.isolator != nil {
			policy := sandboxos.MCPServerPolicyWithProtectedPaths(sc.workspacePath, sc.dataRoot, sc.protectedPaths)
			if err := sc.isolator.Apply(context.Background(), cmd, policy); err != nil {
				if sc.failClosed {
					sc.publishSandboxDecision("rejected", err.Error())
					return nil, fmt.Errorf("%w: MCP server %q: %v", sandboxos.ErrSandboxRequired, sc.name, err)
				}
				log := logging.App()
				log.Warnw("MCP server OS sandbox unavailable", "server", sc.name, "error", err)
				sc.publishSandboxDecision("skipped", err.Error())
			} else {
				sc.publishSandboxDecision("applied", "")
			}
		} else {
			// No isolator configured and fail-open: still record so the
			// audit trail shows MCP servers that ran unsandboxed.
			sc.publishSandboxDecision("skipped", "no isolator configured")
		}

		return &sdkmcp.CommandTransport{Command: cmd}, nil

	case "http":
		if sc.cfg.URL == "" {
			return nil, fmt.Errorf("%w: http requires url", ErrInvalidTransport)
		}
		t := &sdkmcp.StreamableClientTransport{
			Endpoint: sc.cfg.URL,
		}
		if len(sc.cfg.Headers) > 0 {
			t.HTTPClient = &http.Client{
				Transport: &headerRoundTripper{
					base:    http.DefaultTransport,
					headers: sc.cfg.Headers,
				},
			}
		}
		return t, nil

	case "sse":
		if sc.cfg.URL == "" {
			return nil, fmt.Errorf("%w: sse requires url", ErrInvalidTransport)
		}
		t := &sdkmcp.SSEClientTransport{
			Endpoint: sc.cfg.URL,
		}
		if len(sc.cfg.Headers) > 0 {
			t.HTTPClient = &http.Client{
				Transport: &headerRoundTripper{
					base:    http.DefaultTransport,
					headers: sc.cfg.Headers,
				},
			}
		}
		return t, nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidTransport, sc.cfg.Transport)
	}
}

func (sc *ServerConnection) discoverCapabilities(ctx context.Context) {
	session := sc.Session()
	if session == nil {
		return
	}

	discoverCtx, cancel := context.WithTimeout(ctx, sc.timeout())
	defer cancel()

	// Discover tools
	var tools []DiscoveredTool
	for tool, err := range session.Tools(discoverCtx, nil) {
		if err != nil {
			logging.App().Warnw("MCP tool discovery error", "server", sc.name, "error", err)
			break
		}
		tools = append(tools, DiscoveredTool{
			ServerName: sc.name,
			Tool:       tool,
		})
	}

	// Discover resources
	var resources []DiscoveredResource
	for res, err := range session.Resources(discoverCtx, nil) {
		if err != nil {
			logging.App().Debugw("MCP resource discovery error", "server", sc.name, "error", err)
			break
		}
		resources = append(resources, DiscoveredResource{
			ServerName: sc.name,
			Resource:   res,
		})
	}

	// Discover prompts
	var prompts []DiscoveredPrompt
	for p, err := range session.Prompts(discoverCtx, nil) {
		if err != nil {
			logging.App().Debugw("MCP prompt discovery error", "server", sc.name, "error", err)
			break
		}
		prompts = append(prompts, DiscoveredPrompt{
			ServerName: sc.name,
			Prompt:     p,
		})
	}

	sc.mu.Lock()
	sc.tools = tools
	sc.resources = resources
	sc.prompts = prompts
	sc.mu.Unlock()
}

// headerRoundTripper injects HTTP headers into every request.
type headerRoundTripper struct {
	base    http.RoundTripper
	headers map[string]string
}

func (rt *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range rt.headers {
		req.Header.Set(k, v)
	}
	return rt.base.RoundTrip(req)
}
