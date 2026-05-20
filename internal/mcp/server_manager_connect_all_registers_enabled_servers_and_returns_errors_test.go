package mcp

import (
	"context"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
)

func TestServerManagerConnectAllRegistersEnabledServersAndReturnsErrors(t *testing.T) {
	t.Parallel()

	disabled := false
	mgr := NewServerManager(config.MCPConfig{
		Servers: map[string]config.MCPServerConfig{
			"disabled": {
				Transport: "grpc",
				Enabled:   &disabled,
			},
			"invalid": {
				Transport: "grpc",
			},
			"fail-closed": {
				Transport: "stdio",
				Command:   "echo",
			},
		},
	})
	mgr.SetFailClosed(true)

	errs := mgr.ConnectAll(context.Background())

	require.Len(t, errs, 2)
	require.ErrorIs(t, errs["invalid"], ErrConnectionFailed)
	require.Contains(t, errs["invalid"].Error(), `"grpc"`)
	require.ErrorIs(t, errs["fail-closed"], ErrConnectionFailed)
	require.Contains(t, errs["fail-closed"].Error(), "no OS isolator configured")

	require.Equal(t, 2, mgr.ServerCount())
	_, ok := mgr.GetConnection("disabled")
	require.False(t, ok)

	invalid, ok := mgr.GetConnection("invalid")
	require.True(t, ok)
	require.Equal(t, StateFailed, invalid.State())

	failClosed, ok := mgr.GetConnection("fail-closed")
	require.True(t, ok)
	require.Equal(t, StateFailed, failClosed.State())
}

func TestServerManagerSettersPropagateToExistingAndFutureConnections(t *testing.T) {
	t.Parallel()

	mgr := NewServerManager(config.MCPConfig{})
	current := NewServerConnection("current", config.MCPServerConfig{}, config.MCPConfig{})
	mgr.mu.Lock()
	mgr.servers["current"] = current
	mgr.mu.Unlock()

	iso := &recordingIsolator{name: "p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9-sandbox"}
	bus := eventbus.New()
	workspace := t.TempDir()
	dataRoot := t.TempDir()
	protected := []string{t.TempDir()}
	originalProtected := protected[0]

	mgr.SetOSIsolator(iso, workspace, dataRoot)
	mgr.SetProtectedPaths(protected)
	mgr.SetFailClosed(true)
	mgr.SetEventBus(bus)
	protected[0] = t.TempDir()

	current.mu.RLock()
	require.Same(t, iso, current.isolator)
	require.Equal(t, workspace, current.workspacePath)
	require.Equal(t, dataRoot, current.dataRoot)
	require.Equal(t, []string{originalProtected}, current.protectedPaths)
	require.True(t, current.failClosed)
	require.Same(t, bus, current.bus)
	current.mu.RUnlock()

	mgr.cfg.Servers = map[string]config.MCPServerConfig{
		"future": {Transport: "grpc"},
	}
	errs := mgr.ConnectAll(context.Background())
	require.ErrorIs(t, errs["future"], ErrConnectionFailed)

	future, ok := mgr.GetConnection("future")
	require.True(t, ok)
	future.mu.RLock()
	require.Same(t, iso, future.isolator)
	require.Equal(t, workspace, future.workspacePath)
	require.Equal(t, dataRoot, future.dataRoot)
	require.Equal(t, []string{originalProtected}, future.protectedPaths)
	require.True(t, future.failClosed)
	require.Same(t, bus, future.bus)
	future.mu.RUnlock()
}

func TestServerManagerAggregatesCapabilitiesAndStatusFromRegisteredConnections(t *testing.T) {
	t.Parallel()

	alpha := NewServerConnection("alpha", config.MCPServerConfig{}, config.MCPConfig{})
	beta := NewServerConnection("beta", config.MCPServerConfig{}, config.MCPConfig{})

	alpha.mu.Lock()
	alpha.state = StateConnected
	alpha.tools = []DiscoveredTool{{ServerName: "alpha", Tool: &sdkmcp.Tool{Name: "tool-a"}}}
	alpha.resources = []DiscoveredResource{
		{ServerName: "alpha", Resource: &sdkmcp.Resource{Name: "resource-a"}},
	}
	alpha.prompts = []DiscoveredPrompt{{ServerName: "alpha", Prompt: &sdkmcp.Prompt{Name: "prompt-a"}}}
	alpha.mu.Unlock()

	beta.mu.Lock()
	beta.state = StateFailed
	beta.tools = []DiscoveredTool{{ServerName: "beta", Tool: &sdkmcp.Tool{Name: "tool-b"}}}
	beta.resources = []DiscoveredResource{
		{ServerName: "beta", Resource: &sdkmcp.Resource{Name: "resource-b"}},
	}
	beta.prompts = []DiscoveredPrompt{{ServerName: "beta", Prompt: &sdkmcp.Prompt{Name: "prompt-b"}}}
	beta.mu.Unlock()

	mgr := NewServerManager(config.MCPConfig{})
	mgr.mu.Lock()
	mgr.servers["alpha"] = alpha
	mgr.servers["beta"] = beta
	mgr.mu.Unlock()

	require.ElementsMatch(t, []string{"alpha/tool-a", "beta/tool-b"}, serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsToolNames(mgr.AllTools()))
	require.ElementsMatch(t, []string{"alpha/resource-a", "beta/resource-b"},
		serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsResourceNames(mgr.AllResources()))
	require.ElementsMatch(t, []string{"alpha/prompt-a", "beta/prompt-b"},
		serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsPromptNames(mgr.AllPrompts()))
	require.Equal(t, map[string]ServerState{
		"alpha": StateConnected,
		"beta":  StateFailed,
	}, mgr.ServerStatus())
}

func TestServerManagerAggregateResultsAreSliceCopies(t *testing.T) {
	t.Parallel()

	conn := NewServerConnection("copy-source", config.MCPServerConfig{}, config.MCPConfig{})
	conn.mu.Lock()
	conn.tools = []DiscoveredTool{{ServerName: "copy-source", Tool: &sdkmcp.Tool{Name: "tool"}}}
	conn.resources = []DiscoveredResource{
		{ServerName: "copy-source", Resource: &sdkmcp.Resource{Name: "resource"}},
	}
	conn.prompts = []DiscoveredPrompt{
		{ServerName: "copy-source", Prompt: &sdkmcp.Prompt{Name: "prompt"}},
	}
	conn.mu.Unlock()

	mgr := NewServerManager(config.MCPConfig{})
	mgr.mu.Lock()
	mgr.servers["copy-source"] = conn
	mgr.mu.Unlock()

	tools := mgr.AllTools()
	require.Len(t, tools, 1)
	tools[0] = DiscoveredTool{ServerName: "mutated", Tool: &sdkmcp.Tool{Name: "mutated"}}
	require.Equal(t, []string{"copy-source/tool"}, serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsToolNames(mgr.AllTools()))

	resources := mgr.AllResources()
	require.Len(t, resources, 1)
	resources[0] = DiscoveredResource{
		ServerName: "mutated",
		Resource:   &sdkmcp.Resource{Name: "mutated"},
	}
	require.Equal(t, []string{"copy-source/resource"}, serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsResourceNames(mgr.AllResources()))

	prompts := mgr.AllPrompts()
	require.Len(t, prompts, 1)
	prompts[0] = DiscoveredPrompt{ServerName: "mutated", Prompt: &sdkmcp.Prompt{Name: "mutated"}}
	require.Equal(t, []string{"copy-source/prompt"}, serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsPromptNames(mgr.AllPrompts()))
}

func TestServerManagerDisconnectAllStopsRegisteredConnections(t *testing.T) {
	t.Parallel()

	clientSession, serverSession := serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsInMemorySession(t)
	connected := NewServerConnection("connected", config.MCPServerConfig{}, config.MCPConfig{})
	connected.mu.Lock()
	connected.session = clientSession
	connected.state = StateConnected
	connected.mu.Unlock()

	alreadyFailed := NewServerConnection("failed", config.MCPServerConfig{}, config.MCPConfig{})
	alreadyFailed.setState(StateFailed)

	mgr := NewServerManager(config.MCPConfig{})
	mgr.mu.Lock()
	mgr.servers["connected"] = connected
	mgr.servers["failed"] = alreadyFailed
	mgr.mu.Unlock()

	require.NoError(t, mgr.DisconnectAll(context.Background()))
	require.Equal(t, StateStopped, connected.State())
	require.Nil(t, connected.Session())
	require.Equal(t, StateStopped, alreadyFailed.State())
	require.NoError(t, serverSession.Close())
}

func serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsInMemorySession(t *testing.T) (*sdkmcp.ClientSession, *sdkmcp.ServerSession) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9-server", Version: "test"},
		nil,
	)
	client := sdkmcp.NewClient(
		&sdkmcp.Implementation{Name: "p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9-client", Version: "test"},
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

func serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsToolNames(tools []DiscoveredTool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.ServerName+"/"+tool.Tool.Name)
	}
	return names
}

func serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsResourceNames(resources []DiscoveredResource) []string {
	names := make([]string, 0, len(resources))
	for _, resource := range resources {
		names = append(names, resource.ServerName+"/"+resource.Resource.Name)
	}
	return names
}

func serverManagerConnectAllRegistersEnabledServersAndReturnsErrorsPromptNames(prompts []DiscoveredPrompt) []string {
	names := make([]string, 0, len(prompts))
	for _, prompt := range prompts {
		names = append(names, prompt.ServerName+"/"+prompt.Prompt.Name)
	}
	return names
}
