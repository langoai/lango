package mcp

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

func TestServerConnectionToolsReturnsSliceCopy(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("copy-server", config.MCPServerConfig{}, config.MCPConfig{})
	originalTool := &sdkmcp.Tool{Name: "original"}
	replacementTool := &sdkmcp.Tool{Name: "replacement"}

	conn.mu.Lock()
	conn.tools = []DiscoveredTool{{ServerName: "copy-server", Tool: originalTool}}
	conn.mu.Unlock()

	got := conn.Tools()
	require.Len(t, got, 1)
	got[0] = DiscoveredTool{ServerName: "mutated", Tool: replacementTool}
	got = append(got, DiscoveredTool{ServerName: "extra", Tool: replacementTool})

	again := conn.Tools()
	require.Len(t, again, 1)
	require.Equal(t, "copy-server", again[0].ServerName)
	require.Same(t, originalTool, again[0].Tool)
}

func TestServerConnectionStdioTransportPublishesFailOpenSandboxDecision(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var decisions []eventbus.SandboxDecisionEvent
	eventbus.SubscribeTyped(bus, func(event eventbus.SandboxDecisionEvent) {
		decisions = append(decisions, event)
	})

	workspace := t.TempDir()
	conn := NewServerConnection("unsandboxed",
		config.MCPServerConfig{
			Transport: "stdio",
			Command:   "echo",
			Args:      []string{"hello"},
			Env:       map[string]string{"MCP_TEST_ENV": "serverConnectionToolsReturnsSliceCopy2"},
		},
		config.MCPConfig{},
	)
	conn.SetOSIsolator(nil, workspace, t.TempDir())
	conn.SetEventBus(bus)

	transport, err := conn.createTransport()
	require.NoError(t, err)
	commandTransport, ok := transport.(*sdkmcp.CommandTransport)
	require.True(t, ok)
	require.Equal(t, workspace, commandTransport.Command.Dir)
	require.Equal(t, []string{"echo", "hello"}, commandTransport.Command.Args)
	require.Contains(t, commandTransport.Command.Env, "MCP_TEST_ENV=serverConnectionToolsReturnsSliceCopy2")

	require.Len(t, decisions, 1)
	require.Equal(t, "mcp", decisions[0].Source)
	require.Equal(t, "unsandboxed", decisions[0].Command)
	require.Equal(t, "skipped", decisions[0].Decision)
	require.Empty(t, decisions[0].Backend)
	require.Equal(t, "no isolator configured", decisions[0].Reason)
	require.False(t, decisions[0].Timestamp.IsZero())
}

func TestServerConnectionSandboxApplyErrorPublishesDecisionByMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		failClosed   bool
		wantDecision string
		wantErr      bool
	}{
		{name: "fail open publishes skipped", wantDecision: "skipped"},
		{name: "fail closed publishes rejected", failClosed: true, wantDecision: "rejected", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bus := eventbus.New()
			var decisions []eventbus.SandboxDecisionEvent
			eventbus.SubscribeTyped(bus, func(event eventbus.SandboxDecisionEvent) {
				decisions = append(decisions, event)
			})

			conn := NewServerConnection(tt.name,
				config.MCPServerConfig{Transport: "stdio", Command: "echo"},
				config.MCPConfig{},
			)
			conn.SetOSIsolator(&recordingIsolator{
				name: "fake-sandbox",
				err:  errors.New("sandbox unavailable"),
			}, t.TempDir(), "")
			conn.SetFailClosed(tt.failClosed)
			conn.SetEventBus(bus)

			transport, err := conn.createTransport()
			if tt.wantErr {
				require.ErrorIs(t, err, sandboxos.ErrSandboxRequired)
				require.Nil(t, transport)
			} else {
				require.NoError(t, err)
				require.NotNil(t, transport)
			}

			require.Len(t, decisions, 1)
			require.Equal(t, tt.wantDecision, decisions[0].Decision)
			require.Equal(t, "fake-sandbox", decisions[0].Backend)
			require.Equal(t, "sandbox unavailable", decisions[0].Reason)
		})
	}
}

func TestServerConnectionTimeoutPrefersServerThenGlobalThenDefault(t *testing.T) {
	t.Parallel()

	serverTimeout := NewServerConnection("server-timeout",
		config.MCPServerConfig{Timeout: 2 * time.Second},
		config.MCPConfig{DefaultTimeout: 5 * time.Second},
	)
	require.Equal(t, 2*time.Second, serverTimeout.timeout())

	globalTimeout := NewServerConnection("global-timeout",
		config.MCPServerConfig{},
		config.MCPConfig{DefaultTimeout: 5 * time.Second},
	)
	require.Equal(t, 5*time.Second, globalTimeout.timeout())

	defaultTimeout := NewServerConnection("default-timeout",
		config.MCPServerConfig{},
		config.MCPConfig{},
	)
	require.Equal(t, 30*time.Second, defaultTimeout.timeout())
}

func TestServerConnectionConnectTransportStartErrorSetsFailedState(t *testing.T) {
	t.Parallel()

	missingCommand := filepath.Join(t.TempDir(), "missing-mcp-command-serverConnectionToolsReturnsSliceCopy2")
	conn := NewServerConnection("missing-command",
		config.MCPServerConfig{
			Transport: "stdio",
			Command:   missingCommand,
		},
		config.MCPConfig{},
	)

	err := conn.Connect(context.Background())
	require.ErrorIs(t, err, ErrConnectionFailed)
	require.Contains(t, err.Error(), missingCommand)
	require.Equal(t, StateFailed, conn.State())
	require.Nil(t, conn.Session())
}

func TestServerConnectionDisconnectClosesInMemorySession(t *testing.T) {
	t.Parallel()

	clientSession, _ := newServerConnectionToolsReturnsSliceCopyInMemorySession(t)
	conn := NewServerConnection("disconnect", config.MCPServerConfig{}, config.MCPConfig{})
	conn.mu.Lock()
	conn.session = clientSession
	conn.state = StateConnected
	conn.mu.Unlock()

	require.NoError(t, conn.Disconnect(context.Background()))
	require.Equal(t, StateStopped, conn.State())
	require.Nil(t, conn.Session())
	require.NoError(t, conn.Disconnect(context.Background()))
}

func TestServerConnectionHealthCheckUsesInMemorySessionState(t *testing.T) {
	t.Parallel()

	t.Run("keeps connected state after successful ping", func(t *testing.T) {
		t.Parallel()

		clientSession, _ := newServerConnectionToolsReturnsSliceCopyInMemorySession(t)
		conn := NewServerConnection("healthy", config.MCPServerConfig{}, config.MCPConfig{})
		conn.mu.Lock()
		conn.session = clientSession
		conn.state = StateConnected
		conn.mu.Unlock()

		conn.healthCheck(context.Background())

		require.Equal(t, StateConnected, conn.State())
		require.Same(t, clientSession, conn.Session())
	})

	t.Run("marks failed after canceled ping context", func(t *testing.T) {
		t.Parallel()

		clientSession, _ := newServerConnectionToolsReturnsSliceCopyInMemorySession(t)
		conn := NewServerConnection("unhealthy", config.MCPServerConfig{}, config.MCPConfig{})
		conn.mu.Lock()
		conn.session = clientSession
		conn.state = StateConnected
		conn.mu.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		conn.healthCheck(ctx)

		require.Equal(t, StateFailed, conn.State())
		require.Same(t, clientSession, conn.Session())
	})
}

func newServerConnectionToolsReturnsSliceCopyInMemorySession(t *testing.T) (*sdkmcp.ClientSession, *sdkmcp.ServerSession) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "serverConnectionToolsReturnsSliceCopy2-server", Version: "test"},
		nil,
	)
	client := sdkmcp.NewClient(
		&sdkmcp.Implementation{Name: "serverConnectionToolsReturnsSliceCopy2-client", Version: "test"},
		nil,
	)

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = clientSession.Close()
		_ = serverSession.Close()
	})

	return clientSession, serverSession
}
