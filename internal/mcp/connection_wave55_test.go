package mcp

import (
	"context"
	"errors"
	"net/http"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

func TestServerConnectionWave55_StdioTransportBuildsLocalCommandWithoutSandbox(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	conn := NewServerConnection("local-stdio",
		config.MCPServerConfig{
			Transport: "stdio",
			Command:   "echo",
			Args:      []string{"hello", "wave55"},
			Env:       map[string]string{"WAVE": "55"},
		},
		config.MCPConfig{},
	)
	conn.SetOSIsolator(nil, workspace, "")

	transport, err := conn.createTransport()
	require.NoError(t, err)
	commandTransport, ok := transport.(*sdkmcp.CommandTransport)
	require.True(t, ok)
	require.Equal(t, []string{"echo", "hello", "wave55"}, commandTransport.Command.Args)
	require.Equal(t, workspace, commandTransport.Command.Dir)
	require.Contains(t, commandTransport.Command.Env, "WAVE=55")
}

func TestServerConnectionWave55_HTTPAndSSETransportsInjectHeadersDeterministically(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  config.MCPServerConfig
	}{
		{
			name: "http",
			cfg: config.MCPServerConfig{
				Transport: "http",
				URL:       "http://example.invalid/mcp",
				Headers:   map[string]string{"Authorization": "Bearer wave55"},
			},
		},
		{
			name: "sse",
			cfg: config.MCPServerConfig{
				Transport: "sse",
				URL:       "http://example.invalid/sse",
				Headers:   map[string]string{"X-Wave": "55"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn := NewServerConnection("headers", tt.cfg, config.MCPConfig{})
			transport, err := conn.createTransport()
			require.NoError(t, err)

			req, err := wave55RoundTripWithStub(t, transport, tt.cfg.Transport, tt.cfg.URL)
			require.NoError(t, err)
			for key, value := range tt.cfg.Headers {
				require.Equal(t, value, req.Header.Get(key))
			}
		})
	}
}

func wave55RoundTripWithStub(t *testing.T, transport sdkmcp.Transport, transportType, url string) (*http.Request, error) {
	t.Helper()

	stub := &stubRoundTripper{}
	method := http.MethodPost
	switch transportType {
	case "http":
		httpTransport, ok := transport.(*sdkmcp.StreamableClientTransport)
		require.True(t, ok)
		rt, ok := httpTransport.HTTPClient.Transport.(*headerRoundTripper)
		require.True(t, ok)
		rt.base = stub
	case "sse":
		sseTransport, ok := transport.(*sdkmcp.SSEClientTransport)
		require.True(t, ok)
		rt, ok := sseTransport.HTTPClient.Transport.(*headerRoundTripper)
		require.True(t, ok)
		rt.base = stub
		method = http.MethodGet
	default:
		t.Fatalf("unsupported transport type %q", transportType)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	switch transportType {
	case "http":
		httpTransport := transport.(*sdkmcp.StreamableClientTransport)
		_, err = httpTransport.HTTPClient.Transport.RoundTrip(req)
	case "sse":
		sseTransport := transport.(*sdkmcp.SSEClientTransport)
		_, err = sseTransport.HTTPClient.Transport.RoundTrip(req)
	}
	return stub.lastReq, err
}

func TestServerConnectionWave55_FailClosedNilIsolatorPublishesRejectedDecision(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var decisions []eventbus.SandboxDecisionEvent
	eventbus.SubscribeTyped(bus, func(event eventbus.SandboxDecisionEvent) {
		decisions = append(decisions, event)
	})

	conn := NewServerConnection("nil-isolator",
		config.MCPServerConfig{Transport: "stdio", Command: "echo"},
		config.MCPConfig{},
	)
	conn.SetFailClosed(true)
	conn.SetEventBus(bus)

	transport, err := conn.createTransport()
	require.Nil(t, transport)
	require.ErrorIs(t, err, sandboxos.ErrSandboxRequired)
	require.Len(t, decisions, 1)
	require.Equal(t, "rejected", decisions[0].Decision)
	require.Equal(t, "no isolator configured", decisions[0].Reason)
	require.Empty(t, decisions[0].Backend)
}

func TestServerConnectionWave55_ConnectValidationFailureSetsFailedWithoutSession(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("missing-url",
		config.MCPServerConfig{Transport: "http"},
		config.MCPConfig{},
	)

	err := conn.Connect(context.Background())
	require.ErrorIs(t, err, ErrConnectionFailed)
	require.Contains(t, err.Error(), "http requires url")
	require.Equal(t, StateFailed, conn.State())
	require.Nil(t, conn.Session())
}

func TestServerConnectionWave55_FailOpenSandboxApplyErrorPublishesSkippedDecision(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var decisions []eventbus.SandboxDecisionEvent
	eventbus.SubscribeTyped(bus, func(event eventbus.SandboxDecisionEvent) {
		decisions = append(decisions, event)
	})

	conn := NewServerConnection("sandbox-skip",
		config.MCPServerConfig{Transport: "stdio", Command: "echo"},
		config.MCPConfig{},
	)
	conn.SetOSIsolator(&recordingIsolator{name: "fake", err: errors.New("sandbox disabled")}, t.TempDir(), "")
	conn.SetEventBus(bus)

	transport, err := conn.createTransport()
	require.NoError(t, err)
	require.NotNil(t, transport)
	require.Len(t, decisions, 1)
	require.Equal(t, "skipped", decisions[0].Decision)
	require.Equal(t, "fake", decisions[0].Backend)
	require.Equal(t, "sandbox disabled", decisions[0].Reason)
}
