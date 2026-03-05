package agentpool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestPool(t *testing.T) *Pool {
	t.Helper()
	return New(zap.NewNop().Sugar())
}

func TestPoolProvider_AvailableAgents(t *testing.T) {
	pool := newTestPool(t)
	require.NoError(t, pool.Add(&Agent{DID: "did:1", Name: "agent-a", Status: StatusHealthy, Capabilities: []string{"code"}}))
	require.NoError(t, pool.Add(&Agent{DID: "did:2", Name: "agent-b", Status: StatusUnhealthy, Capabilities: []string{"search"}}))
	require.NoError(t, pool.Add(&Agent{DID: "did:3", Name: "agent-c", Status: StatusDegraded, Capabilities: []string{"code", "search"}}))

	provider := NewPoolProvider(pool, nil)
	agents := provider.AvailableAgents()

	// Should exclude unhealthy agents.
	assert.Len(t, agents, 2)

	names := make(map[string]bool)
	for _, a := range agents {
		names[a.Name] = true
	}
	assert.True(t, names["agent-a"])
	assert.True(t, names["agent-c"])
	assert.False(t, names["agent-b"])
}

func TestPoolProvider_FindForCapability(t *testing.T) {
	pool := newTestPool(t)
	require.NoError(t, pool.Add(&Agent{DID: "did:1", Name: "coder", Status: StatusHealthy, Capabilities: []string{"code", "review"}}))
	require.NoError(t, pool.Add(&Agent{DID: "did:2", Name: "searcher", Status: StatusHealthy, Capabilities: []string{"search"}}))
	require.NoError(t, pool.Add(&Agent{DID: "did:3", Name: "all-in-one", Status: StatusHealthy, Capabilities: []string{"code", "search"}}))

	provider := NewPoolProvider(pool, nil)

	tests := []struct {
		name       string
		capability string
		wantCount  int
	}{
		{name: "code capability", capability: "code", wantCount: 2},
		{name: "search capability", capability: "search", wantCount: 2},
		{name: "review capability", capability: "review", wantCount: 1},
		{name: "nonexistent capability", capability: "deploy", wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agents := provider.FindForCapability(tt.capability)
			assert.Len(t, agents, tt.wantCount)
		})
	}
}

func TestPoolProvider_EmptyPool(t *testing.T) {
	pool := newTestPool(t)
	provider := NewPoolProvider(pool, nil)

	assert.Empty(t, provider.AvailableAgents())
	assert.Empty(t, provider.FindForCapability("any"))
}

func TestDynamicAgentInfo_Fields(t *testing.T) {
	pool := newTestPool(t)
	require.NoError(t, pool.Add(&Agent{
		DID:          "did:test:123",
		Name:         "test-agent",
		PeerID:       "peer-abc",
		Status:       StatusHealthy,
		Capabilities: []string{"code"},
		TrustScore:   0.85,
		PricePerCall: 0.01,
	}))

	provider := NewPoolProvider(pool, nil)
	agents := provider.AvailableAgents()
	require.Len(t, agents, 1)

	info := agents[0]
	assert.Equal(t, "test-agent", info.Name)
	assert.Equal(t, "did:test:123", info.DID)
	assert.Equal(t, "peer-abc", info.PeerID)
	assert.Equal(t, []string{"code"}, info.Capabilities)
	assert.Equal(t, 0.85, info.TrustScore)
	assert.Equal(t, 0.01, info.PricePerCall)
}
