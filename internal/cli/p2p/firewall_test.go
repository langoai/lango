package p2p

import (
	"encoding/json"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirewallListCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.FirewallRules = nil

	cmd := newFirewallListCmd(testutil.FakeBootLoader(t, cfg))
	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No firewall rules configured. Default policy: deny-all.")
}

func TestFirewallListCmd_WritesTableToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.FirewallRules = []config.FirewallRule{
		{
			PeerDID:   "did:lango:02abc",
			Action:    "allow",
			Tools:     []string{"search_*", "rag_*"},
			RateLimit: 10,
		},
	}

	cmd := newFirewallListCmd(testutil.FakeBootLoader(t, cfg))
	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "PEER DID")
	assert.Contains(t, out, "did:lango:02abc")
	assert.Contains(t, out, "allow")
	assert.Contains(t, out, "10/min")
}

func TestFirewallListCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.FirewallRules = []config.FirewallRule{
		{
			PeerDID: "did:lango:json",
			Action:  "deny",
		},
	}

	cmd := newFirewallListCmd(testutil.FakeBootLoader(t, cfg))
	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "did:lango:json", decoded[0]["peerDid"])
	assert.Equal(t, "deny", decoded[0]["action"])
}

func TestFirewallAddCmd_WritesToCommandWriter(t *testing.T) {
	cmd := newFirewallAddCmd(dummyBootLoader())
	out, err := executeP2PCmd(t, cmd,
		"--peer-did", "did:lango:add",
		"--action", "allow",
		"--tools", "search_*,rag_*",
		"--rate-limit", "12",
	)
	require.NoError(t, err)
	assert.Contains(t, out, "Firewall rule added (runtime only):")
	assert.Contains(t, out, "Peer DID:    did:lango:add")
	assert.Contains(t, out, "Rate Limit:  12/min")
	assert.Contains(t, out, "To persist this rule, add it to p2p.firewallRules in your configuration.")
}

func TestFirewallRemoveCmd_WritesToCommandWriter(t *testing.T) {
	cmd := newFirewallRemoveCmd(dummyBootLoader())
	out, err := executeP2PCmd(t, cmd, "did:lango:remove")
	require.NoError(t, err)
	assert.Contains(t, out, "To remove rules for peer did:lango:remove, edit p2p.firewallRules in your configuration.")
	assert.Contains(t, out, "Runtime rule removal requires the P2P node to be running via 'lango serve'.")
}
