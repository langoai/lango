package mcp

import (
	"context"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestServerConnectionDiscoverCapabilitiesRecordsServerItems(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "capability-server", Version: "test"},
		nil,
	)
	server.AddTool(&sdkmcp.Tool{
		Name:        "lookup",
		Description: "Lookup data",
		InputSchema: map[string]any{"type": "object"},
	}, nil)
	server.AddResource(&sdkmcp.Resource{
		Name: "readme",
		URI:  "file:///readme.md",
	}, nil)
	server.AddPrompt(&sdkmcp.Prompt{
		Name:        "summarize",
		Description: "Summarize context",
	}, nil)

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	client := sdkmcp.NewClient(
		&sdkmcp.Implementation{Name: "capability-client", Version: "test"},
		nil,
	)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = clientSession.Close()
		_ = serverSession.Close()
	})

	conn := NewServerConnection("local", config.MCPServerConfig{}, config.MCPConfig{})
	conn.mu.Lock()
	conn.session = clientSession
	conn.mu.Unlock()

	conn.discoverCapabilities(ctx)

	tools := conn.Tools()
	require.Len(t, tools, 1)
	require.Equal(t, "local", tools[0].ServerName)
	require.Equal(t, "lookup", tools[0].Tool.Name)
	require.Equal(t, "Lookup data", tools[0].Tool.Description)
	require.Equal(t, map[string]any{"type": "object"}, tools[0].Tool.InputSchema)

	conn.mu.RLock()
	resources := append([]DiscoveredResource(nil), conn.resources...)
	prompts := append([]DiscoveredPrompt(nil), conn.prompts...)
	conn.mu.RUnlock()

	require.Len(t, resources, 1)
	require.Equal(t, "local", resources[0].ServerName)
	require.Equal(t, "readme", resources[0].Resource.Name)
	require.Equal(t, "file:///readme.md", resources[0].Resource.URI)

	require.Len(t, prompts, 1)
	require.Equal(t, "local", prompts[0].ServerName)
	require.Equal(t, "summarize", prompts[0].Prompt.Name)
	require.Equal(t, "Summarize context", prompts[0].Prompt.Description)
}
