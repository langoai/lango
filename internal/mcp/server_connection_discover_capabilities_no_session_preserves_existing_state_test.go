package mcp

import (
	"context"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestServerConnectionDiscoverCapabilitiesNoSessionPreservesExistingState(t *testing.T) {
	t.Parallel()

	tool := &sdkmcp.Tool{Name: "kept"}
	conn := NewServerConnection("no-session", config.MCPServerConfig{}, config.MCPConfig{})
	conn.mu.Lock()
	conn.state = StateDisconnected
	conn.tools = []DiscoveredTool{{ServerName: "no-session", Tool: tool}}
	conn.mu.Unlock()

	conn.discoverCapabilities(context.Background())

	require.Equal(t, StateDisconnected, conn.State())
	require.Nil(t, conn.Session())
	require.Equal(t, []DiscoveredTool{{ServerName: "no-session", Tool: tool}}, conn.Tools())
}
