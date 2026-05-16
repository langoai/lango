package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func TestBuildP2PTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := buildP2PTools(&p2pComponents{})

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{name: "connect requires multiaddr", tool: "p2p_connect", params: map[string]interface{}{}, wantErr: "missing multiaddr parameter"},
		{name: "disconnect requires peer did", tool: "p2p_disconnect", params: map[string]interface{}{}, wantErr: "missing peer_did parameter"},
		{name: "query requires peer did", tool: "p2p_query", params: map[string]interface{}{"tool_name": "search_knowledge"}, wantErr: "missing peer_did parameter"},
		{name: "query requires tool name", tool: "p2p_query", params: map[string]interface{}{"peer_did": "did:lango:abc"}, wantErr: "missing tool_name parameter"},
		{name: "firewall add requires peer did", tool: "p2p_firewall_add", params: map[string]interface{}{"action": "allow"}, wantErr: "missing peer_did parameter"},
		{name: "firewall add requires action", tool: "p2p_firewall_add", params: map[string]interface{}{"peer_did": "did:lango:abc"}, wantErr: "missing action parameter"},
		{name: "firewall remove requires peer did", tool: "p2p_firewall_remove", params: map[string]interface{}{}, wantErr: "missing peer_did parameter"},
		{name: "price query requires peer did", tool: "p2p_price_query", params: map[string]interface{}{"tool_name": "search_knowledge"}, wantErr: "missing peer_did parameter"},
		{name: "price query requires tool name", tool: "p2p_price_query", params: map[string]interface{}{"peer_did": "did:lango:abc"}, wantErr: "missing tool_name parameter"},
		{name: "reputation requires peer did", tool: "p2p_reputation", params: map[string]interface{}{}, wantErr: "missing peer_did parameter"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findP2PTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestBuildP2PPaidInvokeTool_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	pc := &paymentComponents{
		wallet:  newP2PTestWallet(t),
		limiter: p2pTestLimiter{},
		chainID: 84532,
	}
	tools := buildP2PPaidInvokeTool(&p2pComponents{}, pc)
	require.Len(t, tools, 1)

	testCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr string
	}{
		{name: "invoke paid requires peer did", params: map[string]interface{}{"tool_name": "search_knowledge"}, wantErr: "missing peer_did parameter"},
		{name: "invoke paid requires tool name", params: map[string]interface{}{"peer_did": "did:lango:abc"}, wantErr: "missing tool_name parameter"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tools[0].Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func findP2PTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
