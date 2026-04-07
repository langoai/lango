package mcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

func TestToolNameFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		serverName string
		toolName   string
		want       string
	}{
		{serverName: "github", toolName: "create_issue", want: "mcp__github__create_issue"},
		{serverName: "slack", toolName: "send_message", want: "mcp__slack__send_message"},
		{serverName: "my-server", toolName: "do_thing", want: "mcp__my-server__do_thing"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := fmt.Sprintf("mcp__%s__%s", tt.serverName, tt.toolName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServerState_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give ServerState
		want string
	}{
		{give: StateDisconnected, want: "disconnected"},
		{give: StateConnecting, want: "connecting"},
		{give: StateConnected, want: "connected"},
		{give: StateFailed, want: "failed"},
		{give: StateStopped, want: "stopped"},
		{give: ServerState(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestNewServerConnection(t *testing.T) {
	t.Parallel()

	cfg := config.MCPServerConfig{
		Transport: "stdio",
		Command:   "node",
		Args:      []string{"server.js"},
	}
	global := config.MCPConfig{
		DefaultTimeout: 30 * time.Second,
	}

	conn := NewServerConnection("test", cfg, global)

	assert.Equal(t, "test", conn.Name())
	assert.Equal(t, StateDisconnected, conn.State())
	assert.Nil(t, conn.Session())
	assert.Empty(t, conn.Tools())
}

func TestServerConnection_Timeout(t *testing.T) {
	t.Parallel()

	t.Run("uses server timeout when set", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Timeout: 10 * time.Second},
			config.MCPConfig{DefaultTimeout: 30 * time.Second},
		)
		assert.Equal(t, 10*time.Second, conn.timeout())
	})

	t.Run("uses global timeout as fallback", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{},
			config.MCPConfig{DefaultTimeout: 45 * time.Second},
		)
		assert.Equal(t, 45*time.Second, conn.timeout())
	})

	t.Run("uses 30s default when neither set", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{},
			config.MCPConfig{},
		)
		assert.Equal(t, 30*time.Second, conn.timeout())
	})
}

func TestServerConnection_CreateTransport_Errors(t *testing.T) {
	t.Parallel()

	t.Run("stdio without command", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Transport: "stdio"},
			config.MCPConfig{},
		)
		_, err := conn.createTransport()
		assert.ErrorIs(t, err, ErrInvalidTransport)
	})

	t.Run("http without url", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Transport: "http"},
			config.MCPConfig{},
		)
		_, err := conn.createTransport()
		assert.ErrorIs(t, err, ErrInvalidTransport)
	})

	t.Run("sse without url", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Transport: "sse"},
			config.MCPConfig{},
		)
		_, err := conn.createTransport()
		assert.ErrorIs(t, err, ErrInvalidTransport)
	})

	t.Run("unknown transport", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Transport: "grpc"},
			config.MCPConfig{},
		)
		_, err := conn.createTransport()
		assert.ErrorIs(t, err, ErrInvalidTransport)
	})
}

func TestServerConnection_CreateTransport_Success(t *testing.T) {
	t.Parallel()

	t.Run("stdio with command", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Transport: "stdio", Command: "echo"},
			config.MCPConfig{},
		)
		transport, err := conn.createTransport()
		assert.NoError(t, err)
		assert.NotNil(t, transport)
	})

	t.Run("http with url", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Transport: "http", URL: "http://localhost:3000"},
			config.MCPConfig{},
		)
		transport, err := conn.createTransport()
		assert.NoError(t, err)
		assert.NotNil(t, transport)
	})

	t.Run("sse with url", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Transport: "sse", URL: "http://localhost:3000/sse"},
			config.MCPConfig{},
		)
		transport, err := conn.createTransport()
		assert.NoError(t, err)
		assert.NotNil(t, transport)
	})

	t.Run("default transport (empty) with command", func(t *testing.T) {
		t.Parallel()
		conn := NewServerConnection("test",
			config.MCPServerConfig{Transport: "", Command: "echo"},
			config.MCPConfig{},
		)
		transport, err := conn.createTransport()
		assert.NoError(t, err)
		assert.NotNil(t, transport)
	})
}

func TestServerManager_Empty(t *testing.T) {
	t.Parallel()

	mgr := NewServerManager(config.MCPConfig{})
	assert.Equal(t, 0, mgr.ServerCount())
	assert.Empty(t, mgr.AllTools())
	assert.Empty(t, mgr.ServerStatus())
}

func TestServerManager_GetConnection_NotFound(t *testing.T) {
	t.Parallel()

	mgr := NewServerManager(config.MCPConfig{})
	_, ok := mgr.GetConnection("nonexistent")
	assert.False(t, ok)
}

func TestServerConnection_SetState(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("test",
		config.MCPServerConfig{},
		config.MCPConfig{},
	)
	assert.Equal(t, StateDisconnected, conn.State())

	conn.setState(StateFailed)
	assert.Equal(t, StateFailed, conn.State())

	conn.setState(StateConnected)
	assert.Equal(t, StateConnected, conn.State())
}

func TestServerConnection_CreateTransport_StdioWithEnv(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("test",
		config.MCPServerConfig{
			Transport: "stdio",
			Command:   "echo",
			Args:      []string{"hello"},
			Env:       map[string]string{"FOO": "bar"},
		},
		config.MCPConfig{},
	)
	transport, err := conn.createTransport()
	assert.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestServerConnection_CreateTransport_HTTPWithHeaders(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("test",
		config.MCPServerConfig{
			Transport: "http",
			URL:       "http://localhost:3000",
			Headers:   map[string]string{"Authorization": "Bearer tok"},
		},
		config.MCPConfig{},
	)
	transport, err := conn.createTransport()
	assert.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestServerConnection_CreateTransport_SSEWithHeaders(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("test",
		config.MCPServerConfig{
			Transport: "sse",
			URL:       "http://localhost:3000/sse",
			Headers:   map[string]string{"X-Key": "val"},
		},
		config.MCPConfig{},
	)
	transport, err := conn.createTransport()
	assert.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestServerManager_AllResources_Empty(t *testing.T) {
	t.Parallel()

	mgr := NewServerManager(config.MCPConfig{})
	assert.Empty(t, mgr.AllResources())
}

func TestServerManager_AllPrompts_Empty(t *testing.T) {
	t.Parallel()

	mgr := NewServerManager(config.MCPConfig{})
	assert.Empty(t, mgr.AllPrompts())
}

// stubRoundTripper records the last request without performing any I/O.
type stubRoundTripper struct {
	lastReq *http.Request
}

func (s *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	s.lastReq = req
	return &http.Response{StatusCode: 200, Body: http.NoBody, Request: req}, nil
}

func TestHeaderRoundTripper(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		"Authorization": "Bearer test-token",
		"X-Custom":      "custom-value",
	}
	stub := &stubRoundTripper{}
	rt := &headerRoundTripper{
		base:    stub,
		headers: headers,
	}

	req, err := http.NewRequest("GET", "http://example.invalid/test", nil)
	require.NoError(t, err)

	_, err = rt.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, "Bearer test-token", stub.lastReq.Header.Get("Authorization"))
	assert.Equal(t, "custom-value", stub.lastReq.Header.Get("X-Custom"))
}

// mockIsolator is a test double for sandboxos.OSIsolator.
type mockIsolator struct {
	err error
}

func (m *mockIsolator) Apply(_ context.Context, _ *exec.Cmd, _ sandboxos.Policy) error {
	return m.err
}

func (m *mockIsolator) Available() bool { return m.err == nil }
func (m *mockIsolator) Name() string    { return "mock" }
func (m *mockIsolator) Reason() string  { return "" }

func TestServerConnection_FailClosed_NilIsolator_Stdio(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("secure-server",
		config.MCPServerConfig{Transport: "stdio", Command: "echo"},
		config.MCPConfig{},
	)
	conn.SetFailClosed(true)

	_, err := conn.createTransport()
	require.Error(t, err)
	assert.ErrorIs(t, err, sandboxos.ErrSandboxRequired)
	assert.Contains(t, err.Error(), "secure-server")
}

func TestServerConnection_FailClosed_ApplyError_Stdio(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("apply-fail",
		config.MCPServerConfig{Transport: "stdio", Command: "echo"},
		config.MCPConfig{},
	)
	conn.SetFailClosed(true)
	conn.SetOSIsolator(&mockIsolator{err: errors.New("landlock unsupported")}, "")

	_, err := conn.createTransport()
	require.Error(t, err)
	assert.ErrorIs(t, err, sandboxos.ErrSandboxRequired)
	assert.Contains(t, err.Error(), "apply-fail")
	assert.Contains(t, err.Error(), "landlock unsupported")
}

func TestServerConnection_FailOpen_ApplyError(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("lenient-server",
		config.MCPServerConfig{Transport: "stdio", Command: "echo"},
		config.MCPConfig{},
	)
	conn.SetOSIsolator(&mockIsolator{err: errors.New("not supported")}, "")

	transport, err := conn.createTransport()
	assert.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestServerConnection_FailClosed_Http(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("http-server",
		config.MCPServerConfig{Transport: "http", URL: "http://localhost:8080"},
		config.MCPConfig{},
	)
	conn.SetFailClosed(true)

	transport, err := conn.createTransport()
	assert.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestServerManager_SetFailClosed_Propagates(t *testing.T) {
	t.Parallel()

	mgr := NewServerManager(config.MCPConfig{
		Servers: map[string]config.MCPServerConfig{
			"s1": {Transport: "stdio", Command: "echo", Enabled: boolPtr(true)},
			"s2": {Transport: "stdio", Command: "echo", Enabled: boolPtr(true)},
		},
	})

	// Manually add connections (without connecting to a real server).
	c1 := NewServerConnection("s1",
		config.MCPServerConfig{Transport: "stdio", Command: "echo"},
		config.MCPConfig{},
	)
	c2 := NewServerConnection("s2",
		config.MCPServerConfig{Transport: "stdio", Command: "echo"},
		config.MCPConfig{},
	)

	mgr.mu.Lock()
	mgr.servers["s1"] = c1
	mgr.servers["s2"] = c2
	mgr.mu.Unlock()

	mgr.SetFailClosed(true)

	// Both connections should now have failClosed=true.
	c1.mu.RLock()
	assert.True(t, c1.failClosed)
	c1.mu.RUnlock()

	c2.mu.RLock()
	assert.True(t, c2.failClosed)
	c2.mu.RUnlock()

	// Verify transport creation is blocked for stdio.
	_, err := c1.createTransport()
	assert.ErrorIs(t, err, sandboxos.ErrSandboxRequired)
}

func boolPtr(b bool) *bool { return &b }
