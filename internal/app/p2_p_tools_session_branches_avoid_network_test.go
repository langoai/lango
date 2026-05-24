package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
)

func TestP2PToolsSessionBranchesAvoidNetwork(t *testing.T) {
	t.Parallel()

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	verified, err := sessions.Create("did:lango:runEnforcesMaxTurnsAfterWrapUpBudget7-verified", true)
	require.NoError(t, err)
	unverified, err := sessions.Create("did:lango:runEnforcesMaxTurnsAfterWrapUpBudget7-unverified", false)
	require.NoError(t, err)

	tools := buildP2PTools(&p2pComponents{
		sessions: sessions,
		fw:       firewall.New(nil, zap.NewNop().Sugar()),
	})

	peers, err := findP2PTool(t, tools, "p2p_peers").Handler(context.Background(), nil)
	require.NoError(t, err)
	peerPayload := p2PToolsMetadataAndMissingDependencyBranchesP2PPayload(t, peers)
	assert.Equal(t, 2, peerPayload["count"])
	peerRows, ok := peerPayload["peers"].([]map[string]interface{})
	require.True(t, ok)
	assert.ElementsMatch(t, []map[string]interface{}{
		{
			"peerDID":    verified.PeerDID,
			"zkVerified": true,
			"createdAt":  verified.CreatedAt.Format(time.RFC3339),
			"expiresAt":  verified.ExpiresAt.Format(time.RFC3339),
		},
		{
			"peerDID":    unverified.PeerDID,
			"zkVerified": false,
			"createdAt":  unverified.CreatedAt.Format(time.RFC3339),
			"expiresAt":  unverified.ExpiresAt.Format(time.RFC3339),
		},
	}, peerRows)

	disconnected, err := findP2PTool(t, tools, "p2p_disconnect").Handler(context.Background(), map[string]interface{}{
		"peer_did": verified.PeerDID,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"status": "disconnected", "peerDID": verified.PeerDID}, disconnected)
	assert.Nil(t, sessions.Get(verified.PeerDID))
	assert.NotNil(t, sessions.Get(unverified.PeerDID))
}

func TestP2PToolsRejectMissingSessionsBeforePeerWork(t *testing.T) {
	t.Parallel()

	tools := buildP2PTools(&p2pComponents{
		sessions: p2PToolsMetadataAndMissingDependencyBranchesP2PSessions(t),
		fw:       firewall.New(nil, zap.NewNop().Sugar()),
	})

	tests := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name: "query stops before DID parsing and remote agent setup",
			tool: "p2p_query",
			params: map[string]interface{}{
				"peer_did":  "not-a-did",
				"tool_name": "search_knowledge",
			},
			wantErr: "no active session for peer not-a-did",
		},
		{
			name: "price query stops before DID parsing and remote agent setup",
			tool: "p2p_price_query",
			params: map[string]interface{}{
				"peer_did":  "not-a-did",
				"tool_name": "search_knowledge",
			},
			wantErr: "no active session for peer not-a-did",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := findP2PTool(t, tools, tt.tool).Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestP2PToolsFirewallMutationAndFormatting(t *testing.T) {
	t.Parallel()

	fw := firewall.New([]firewall.ACLRule{{
		PeerDID:   "did:lango:existing",
		Action:    firewall.ACLActionDeny,
		Tools:     []string{"dangerous_tool"},
		RateLimit: 3,
	}}, zap.NewNop().Sugar())
	tools := buildP2PTools(&p2pComponents{
		sessions: p2PToolsMetadataAndMissingDependencyBranchesP2PSessions(t),
		fw:       fw,
	})

	added, err := findP2PTool(t, tools, "p2p_firewall_add").Handler(context.Background(), map[string]interface{}{
		"peer_did":   "did:lango:runEnforcesMaxTurnsAfterWrapUpBudget7-peer",
		"action":     "allow",
		"tools":      []interface{}{"search_knowledge", "read_memory"},
		"rate_limit": float64(7),
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"status":  "added",
		"message": "Firewall rule added: allow did:lango:runEnforcesMaxTurnsAfterWrapUpBudget7-peer",
	}, added)

	listed, err := findP2PTool(t, tools, "p2p_firewall_rules").Handler(context.Background(), nil)
	require.NoError(t, err)
	payload := p2PToolsMetadataAndMissingDependencyBranchesP2PPayload(t, listed)
	assert.Equal(t, 2, payload["count"])
	rules, ok := payload["rules"].([]map[string]interface{})
	require.True(t, ok)
	assert.ElementsMatch(t, []map[string]interface{}{
		{
			"peerDID":   "did:lango:existing",
			"action":    firewall.ACLActionDeny,
			"tools":     []string{"dangerous_tool"},
			"rateLimit": 3,
		},
		{
			"peerDID":   "did:lango:runEnforcesMaxTurnsAfterWrapUpBudget7-peer",
			"action":    firewall.ACLActionAllow,
			"tools":     []string{"search_knowledge", "read_memory"},
			"rateLimit": 7,
		},
	}, rules)

	removed, err := findP2PTool(t, tools, "p2p_firewall_remove").Handler(context.Background(), map[string]interface{}{
		"peer_did": "did:lango:runEnforcesMaxTurnsAfterWrapUpBudget7-peer",
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"status":  "removed",
		"count":   1,
		"message": "Removed 1 rules for did:lango:runEnforcesMaxTurnsAfterWrapUpBudget7-peer",
	}, removed)
	assert.Len(t, fw.Rules(), 1)
}

func TestP2PToolsUnavailableDependenciesReturnDeterministicResponses(t *testing.T) {
	t.Parallel()

	tools := buildP2PTools(&p2pComponents{
		sessions: p2PToolsMetadataAndMissingDependencyBranchesP2PSessions(t),
		fw:       firewall.New(nil, zap.NewNop().Sugar()),
	})

	discovered, err := findP2PTool(t, tools, "p2p_discover").Handler(context.Background(), map[string]interface{}{
		"capability": "search",
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"peers":   []interface{}{},
		"count":   0,
		"message": "gossip not enabled",
	}, discovered)

	reputation, err := findP2PTool(t, tools, "p2p_reputation").Handler(context.Background(), map[string]interface{}{
		"peer_did": "did:lango:runEnforcesMaxTurnsAfterWrapUpBudget7-peer",
	})
	require.Error(t, err)
	assert.Nil(t, reputation)
	assert.EqualError(t, err, "reputation system not available (requires database)")
}

func TestP2PWiringShortCircuitsBeforeListeners(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = false
	//nolint:staticcheck // intentional: regression test pins legacy KeyDir behavior.
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "disabled-keys")
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}

	assert.Nil(t, initP2P(cfg, &wiringP2PWallet{}, nil, nil, nil, nil, nil, nil, ""))
	//nolint:staticcheck // intentional: regression test pins legacy KeyDir behavior.
	assert.NoDirExists(t, cfg.P2P.KeyDir)

	cfg.P2P.Enabled = true
	//nolint:staticcheck // intentional: regression test pins legacy KeyDir behavior.
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "missing-wallet-keys")
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	//nolint:staticcheck // intentional: regression test pins legacy KeyDir behavior.
	assert.NoDirExists(t, cfg.P2P.KeyDir)
}
