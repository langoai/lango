package mcp

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
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

func TestHeaderRoundTripper(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		"Authorization": "Bearer test-token",
		"X-Custom":      "custom-value",
	}
	rt := &headerRoundTripper{
		base:    http.DefaultTransport,
		headers: headers,
	}

	// Build a request that goes to a non-routable address so it fails,
	// but we can still verify headers were set before the transport call.
	req, err := http.NewRequest("GET", "http://192.0.2.1:1/test", nil)
	require.NoError(t, err)

	// The RoundTrip will fail (connection refused), but headers should be set.
	_, _ = rt.RoundTrip(req)

	assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
	assert.Equal(t, "custom-value", req.Header.Get("X-Custom"))
}
