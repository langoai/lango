package p2p

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/p2p/discovery"
	"github.com/langoai/lango/internal/p2p/handshake"
	p2preputation "github.com/langoai/lango/internal/p2p/reputation"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeP2PCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// dummyBootLoader returns a boot loader that always errors.
// Used for testing command structure without actually bootstrapping.
func dummyBootLoader() func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	}
}

func TestNewP2PCmd_Structure(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	require.NotNil(t, cmd)

	assert.Equal(t, "p2p", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	// Verify all expected subcommands exist.
	expected := []string{
		"status", "peers", "connect", "disconnect",
		"firewall", "discover", "identity", "reputation",
		"pricing", "session", "sandbox", "team", "zkp",
		"provenance", "workspace", "git",
	}

	subCmds := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subCmds[strings.Fields(sub.Use)[0]] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestNewP2PCmd_SubcommandCount(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	assert.Equal(t, 16, len(cmd.Commands()), "expected 16 P2P subcommands")
}

func TestStatusCmd_HasOutputFlag(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "status" {
			outputFlag := sub.Flags().Lookup("output")
			assert.NotNil(t, outputFlag, "status command should have --output flag")
			return
		}
	}
	t.Fatal("status subcommand not found")
}

func TestStatusCmd_WritesTextToCommandWriter(t *testing.T) {
	original := loadStatusCommandData
	loadStatusCommandData = func(_ *bootstrap.Result) (statusCommandData, func(), error) {
		return statusCommandData{
			peerID:         "12D3KooWStatusPeer",
			listenAddrs:    []string{"/ip4/127.0.0.1/tcp/9000"},
			connectedPeers: 3,
			maxPeers:       50,
			mdns:           true,
			relay:          false,
			zkHandshake:    true,
		}, func() {}, nil
	}
	t.Cleanup(func() { loadStatusCommandData = original })

	cmd := newStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "P2P Node Status")
	assert.Contains(t, out, "Peer ID:          12D3KooWStatusPeer")
	assert.Contains(t, out, "Connected Peers:  3 / 50")
	assert.Contains(t, out, "ZK Handshake:     true")
}

func TestStatusCmd_WritesJSONToCommandWriter(t *testing.T) {
	original := loadStatusCommandData
	loadStatusCommandData = func(_ *bootstrap.Result) (statusCommandData, func(), error) {
		return statusCommandData{
			peerID:         "12D3KooWJsonStatus",
			listenAddrs:    []string{"/ip4/127.0.0.1/tcp/9000", "/ip6/::/tcp/9000"},
			connectedPeers: 0,
			maxPeers:       25,
			mdns:           false,
			relay:          true,
			zkHandshake:    false,
		}, func() {}, nil
	}
	t.Cleanup(func() { loadStatusCommandData = original })

	cmd := newStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "12D3KooWJsonStatus", decoded["peerId"])
	assert.Equal(t, float64(0), decoded["connectedPeers"])
	assert.Equal(t, float64(25), decoded["maxPeers"])
	assert.Equal(t, true, decoded["relay"])
}

func TestPeersCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	original := loadPeersCommandData
	loadPeersCommandData = func(_ *bootstrap.Result) ([]peersCommandInfo, func(), error) {
		return []peersCommandInfo{}, func() {}, nil
	}
	t.Cleanup(func() { loadPeersCommandData = original })

	cmd := newPeersCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No connected peers.")
}

func TestPeersCmd_WritesTableToCommandWriter(t *testing.T) {
	original := loadPeersCommandData
	loadPeersCommandData = func(_ *bootstrap.Result) ([]peersCommandInfo, func(), error) {
		return []peersCommandInfo{
			{PeerID: "12D3KooWPeerOne", Addrs: []string{"/ip4/192.168.0.10/tcp/9000"}},
			{PeerID: "12D3KooWPeerTwo", Addrs: []string{}},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadPeersCommandData = original })

	cmd := newPeersCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "PEER ID")
	assert.Contains(t, out, "12D3KooWPeerOne")
	assert.Contains(t, out, "/ip4/192.168.0.10/tcp/9000")
	assert.Contains(t, out, "12D3KooWPeerTwo")
}

func TestPeersCmd_WritesJSONToCommandWriter(t *testing.T) {
	original := loadPeersCommandData
	loadPeersCommandData = func(_ *bootstrap.Result) ([]peersCommandInfo, func(), error) {
		return []peersCommandInfo{
			{PeerID: "12D3KooWPeerJson", Addrs: []string{"/ip4/127.0.0.1/tcp/9000"}},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadPeersCommandData = original })

	cmd := newPeersCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "12D3KooWPeerJson", decoded[0]["peerId"])
}

func TestP2PInspectionCommands_InvalidOutputFailFast(t *testing.T) {
	testCases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{
			name: "status",
			cmd: newStatusCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "peers",
			cmd: newPeersCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "identity",
			cmd: newIdentityCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "discover",
			cmd: newDiscoverCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "pricing",
			cmd: newPricingCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "reputation",
			cmd: newReputationCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "session-list",
			cmd: newSessionListCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "firewall-list",
			cmd: newFirewallListCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "zkp-status",
			cmd: newZKPStatusCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "zkp-circuits",
			cmd:  newZKPCircuitsCmd(),
		},
		{
			name: "team-list",
			cmd: newTeamListCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "team-status",
			cmd: newTeamStatusCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "workspace-create",
			cmd: newWorkspaceCreateCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "workspace-list",
			cmd: newWorkspaceListCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "workspace-status",
			cmd: newWorkspaceStatusCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
		{
			name: "git-log",
			cmd: newGitLogCmd(func() (*bootstrap.Result, error) {
				return nil, assert.AnError
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"--output", "yaml"}
			switch tc.name {
			case "reputation":
				args = append([]string{"--peer-did", "did:lango:test"}, args...)
			case "team-status":
				args = append([]string{"team-123"}, args...)
			case "workspace-create":
				args = append([]string{"workspace-123"}, args...)
			case "workspace-status":
				args = append([]string{"workspace-123"}, args...)
			case "git-log":
				args = append([]string{"workspace-123"}, args...)
			}

			out, err := executeP2PCmd(t, tc.cmd, args...)
			require.EqualError(t, err, `unknown output format "yaml" (expected: table or json)`)
			assert.Equal(t, "", strings.TrimSpace(out))
		})
	}
}

func TestDiscoverCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	original := loadDiscoverCommandData
	loadDiscoverCommandData = func(_ *bootstrap.Result, _ string) ([]*discovery.GossipCard, func(), error) {
		return []*discovery.GossipCard{}, func() {}, nil
	}
	t.Cleanup(func() { loadDiscoverCommandData = original })

	cmd := newDiscoverCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No agents discovered. Try connecting to bootstrap peers first.")
}

func TestDiscoverCmd_WritesTableToCommandWriter(t *testing.T) {
	original := loadDiscoverCommandData
	loadDiscoverCommandData = func(_ *bootstrap.Result, tag string) ([]*discovery.GossipCard, func(), error) {
		require.Equal(t, "research", tag)
		return []*discovery.GossipCard{
			{Name: "research-bot", DID: "did:lango:02abc", Capabilities: []string{"research", "summarize"}, PeerID: "12D3KooWDiscover"},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadDiscoverCommandData = original })

	cmd := newDiscoverCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--tag", "research")
	require.NoError(t, err)
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "research-bot")
	assert.Contains(t, out, "research, summarize")
	assert.Contains(t, out, "12D3KooWDiscover")
}

func TestDiscoverCmd_WritesJSONToCommandWriter(t *testing.T) {
	original := loadDiscoverCommandData
	loadDiscoverCommandData = func(_ *bootstrap.Result, _ string) ([]*discovery.GossipCard, func(), error) {
		return []*discovery.GossipCard{
			{Name: "json-bot", DID: "did:lango:02json", Capabilities: []string{"analyze"}, PeerID: "12D3KooWJsonDiscover"},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadDiscoverCommandData = original })

	cmd := newDiscoverCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "json-bot", decoded[0]["name"])
	assert.Equal(t, "12D3KooWJsonDiscover", decoded[0]["peerId"])
}

func TestReputationCmd_WritesMissingRecordToCommandWriter(t *testing.T) {
	original := loadReputationDetails
	loadReputationDetails = func(_ *bootstrap.Result, peerDID string) (*p2preputation.PeerDetails, error) {
		require.Equal(t, "did:lango:missing", peerDID)
		return nil, nil
	}
	t.Cleanup(func() { loadReputationDetails = original })

	cmd := newReputationCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--peer-did", "did:lango:missing")
	require.NoError(t, err)
	assert.Contains(t, out, "No reputation record found for did:lango:missing")
}

func TestReputationCmd_WritesTextToCommandWriter(t *testing.T) {
	original := loadReputationDetails
	loadReputationDetails = func(_ *bootstrap.Result, _ string) (*p2preputation.PeerDetails, error) {
		return &p2preputation.PeerDetails{
			PeerDID:             "did:lango:peer123",
			TrustScore:          0.8123,
			SuccessfulExchanges: 8,
			FailedExchanges:     1,
			TimeoutCount:        2,
			FirstSeen:           time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
			LastInteraction:     time.Date(2026, 5, 14, 9, 0, 0, 0, time.UTC),
		}, nil
	}
	t.Cleanup(func() { loadReputationDetails = original })

	cmd := newReputationCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--peer-did", "did:lango:peer123")
	require.NoError(t, err)
	assert.Contains(t, out, "Peer Reputation")
	assert.Contains(t, out, "Peer DID:          did:lango:peer123")
	assert.Contains(t, out, "Trust Score:       0.8123")
	assert.Contains(t, out, "Successes:         8")
}

func TestReputationCmd_WritesJSONToCommandWriter(t *testing.T) {
	original := loadReputationDetails
	loadReputationDetails = func(_ *bootstrap.Result, _ string) (*p2preputation.PeerDetails, error) {
		return &p2preputation.PeerDetails{
			PeerDID:             "did:lango:json-peer",
			TrustScore:          0.5,
			SuccessfulExchanges: 5,
			FailedExchanges:     2,
			TimeoutCount:        1,
			FirstSeen:           time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
			LastInteraction:     time.Date(2026, 5, 14, 9, 0, 0, 0, time.UTC),
		}, nil
	}
	t.Cleanup(func() { loadReputationDetails = original })

	cmd := newReputationCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--peer-did", "did:lango:json-peer", "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "did:lango:json-peer", decoded["peerDid"])
	assert.Equal(t, float64(0.5), decoded["trustScore"])
	assert.Equal(t, float64(5), decoded["successfulExchanges"])
}

func TestSessionListCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	original := loadSessionListData
	loadSessionListData = func(_ *bootstrap.Result) ([]*handshake.Session, func(), error) {
		return []*handshake.Session{}, func() {}, nil
	}
	t.Cleanup(func() { loadSessionListData = original })

	cmd := newSessionListCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No active sessions.")
}

func TestSessionListCmd_WritesTableToCommandWriter(t *testing.T) {
	original := loadSessionListData
	loadSessionListData = func(_ *bootstrap.Result) ([]*handshake.Session, func(), error) {
		return []*handshake.Session{
			{
				PeerDID:    "did:lango:peer1",
				CreatedAt:  time.Date(2026, 5, 14, 9, 0, 0, 0, time.UTC),
				ExpiresAt:  time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC),
				ZKVerified: true,
			},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadSessionListData = original })

	cmd := newSessionListCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "PEER DID")
	assert.Contains(t, out, "did:lango:peer1")
	assert.Contains(t, out, "true")
}

func TestSessionListCmd_WritesJSONToCommandWriter(t *testing.T) {
	original := loadSessionListData
	loadSessionListData = func(_ *bootstrap.Result) ([]*handshake.Session, func(), error) {
		return []*handshake.Session{
			{PeerDID: "did:lango:json-peer", ZKVerified: false},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadSessionListData = original })

	cmd := newSessionListCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "did:lango:json-peer", decoded[0]["peerDid"])
}

func TestSessionRevokeCmd_WritesToCommandWriter(t *testing.T) {
	original := revokeSessionForPeer
	revokeSessionForPeer = func(_ *bootstrap.Result, peerDID string) (func(), error) {
		require.Equal(t, "did:lango:revoke", peerDID)
		return func() {}, nil
	}
	t.Cleanup(func() { revokeSessionForPeer = original })

	cmd := newSessionRevokeCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--peer-did", "did:lango:revoke")
	require.NoError(t, err)
	assert.Contains(t, out, "Session for did:lango:revoke revoked.")
}

func TestSessionRevokeAllCmd_WritesToCommandWriter(t *testing.T) {
	original := revokeAllSessions
	revokeAllSessions = func(_ *bootstrap.Result) (func(), error) {
		return func() {}, nil
	}
	t.Cleanup(func() { revokeAllSessions = original })

	cmd := newSessionRevokeAllCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "All sessions revoked.")
}

func TestConnectCmd_WritesToCommandWriter(t *testing.T) {
	original := connectToPeer
	connectToPeer = func(_ context.Context, _ *bootstrap.Result, target string) (string, func(), error) {
		require.Equal(t, "/ip4/192.168.1.5/tcp/9000/p2p/12D3KooWConnect", target)
		return "12D3KooWConnect", func() {}, nil
	}
	t.Cleanup(func() { connectToPeer = original })

	cmd := newConnectCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "/ip4/192.168.1.5/tcp/9000/p2p/12D3KooWConnect")
	require.NoError(t, err)
	assert.Contains(t, out, "Connected to peer 12D3KooWConnect")
}

func TestConnectCmd_PassesCommandContextToConnect(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	cancel()

	original := connectToPeer
	connectToPeer = func(ctx context.Context, _ *bootstrap.Result, _ string) (string, func(), error) {
		require.ErrorIs(t, ctx.Err(), context.Canceled)
		return "", nil, ctx.Err()
	}
	t.Cleanup(func() { connectToPeer = original })

	cmd := newConnectCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})
	cmd.SetContext(parent)

	_, err := executeP2PCmd(t, cmd, validConnectMultiaddr())
	require.ErrorIs(t, err, context.Canceled)
}

func TestConnectToPeer_UsesConfiguredHandshakeTimeout(t *testing.T) {
	var gotDeadline time.Time
	restoreConnectDeps(t, connectDeps{
		config: connectTestP2PConfig(75 * time.Millisecond),
		connect: func(ctx context.Context, _ peer.AddrInfo) error {
			var ok bool
			gotDeadline, ok = ctx.Deadline()
			require.True(t, ok, "connect context should have a deadline")
			return context.DeadlineExceeded
		},
		cleanup: func() {},
	})

	start := time.Now()
	_, _, err := connectToPeer(context.Background(), &bootstrap.Result{}, validConnectMultiaddr())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out after 75ms")
	assert.Contains(t, err.Error(), validConnectPeerID())
	assert.WithinDuration(t, start.Add(75*time.Millisecond), gotDeadline, 250*time.Millisecond)
}

func TestConnectToPeer_FallsBackToDefaultTimeout(t *testing.T) {
	testCases := []struct {
		name    string
		timeout time.Duration
	}{
		{name: "zero", timeout: 0},
		{name: "negative", timeout: -time.Second},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var gotDeadline time.Time
			restoreConnectDeps(t, connectDeps{
				config: connectTestP2PConfig(tc.timeout),
				connect: func(ctx context.Context, _ peer.AddrInfo) error {
					var ok bool
					gotDeadline, ok = ctx.Deadline()
					require.True(t, ok, "connect context should have a deadline")
					return errors.New("dial failed")
				},
				cleanup: func() {},
			})

			start := time.Now()
			_, _, err := connectToPeer(context.Background(), &bootstrap.Result{}, validConnectMultiaddr())
			require.Error(t, err)
			assert.WithinDuration(t, start.Add(30*time.Second), gotDeadline, 250*time.Millisecond)
		})
	}
}

func TestConnectToPeer_CommandCancellationReachesHostConnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	restoreConnectDeps(t, connectDeps{
		config: connectTestP2PConfig(time.Second),
		connect: func(got context.Context, _ peer.AddrInfo) error {
			cancel()
			<-got.Done()
			return got.Err()
		},
		cleanup: func() {},
	})

	_, _, err := connectToPeer(ctx, &bootstrap.Result{}, validConnectMultiaddr())
	require.ErrorIs(t, err, context.Canceled)
	assert.Contains(t, err.Error(), "canceled")
	assert.Contains(t, err.Error(), validConnectPeerID())
}

func TestConnectToPeer_ReportsParentDeadlineSeparatelyFromConfiguredTimeout(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Millisecond))
	defer cancel()

	restoreConnectDeps(t, connectDeps{
		config: connectTestP2PConfig(30 * time.Second),
		connect: func(got context.Context, _ peer.AddrInfo) error {
			<-got.Done()
			return got.Err()
		},
		cleanup: func() {},
	})

	_, _, err := connectToPeer(ctx, &bootstrap.Result{}, validConnectMultiaddr())
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Contains(t, err.Error(), "timed out by command context deadline")
	assert.Contains(t, err.Error(), validConnectPeerID())
	assert.NotContains(t, err.Error(), "timed out after 30s")
}

func TestConnectToPeer_ReportsConfiguredTimeoutWhenItExpiresBeforeParentDeadline(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(25*time.Millisecond))
	defer cancel()

	restoreConnectDeps(t, connectDeps{
		config: connectTestP2PConfig(10 * time.Millisecond),
		connect: func(got context.Context, _ peer.AddrInfo) error {
			<-got.Done()
			time.Sleep(20 * time.Millisecond)
			return got.Err()
		},
		cleanup: func() {},
	})

	_, _, err := connectToPeer(ctx, &bootstrap.Result{}, validConnectMultiaddr())
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Contains(t, err.Error(), "timed out after 10ms")
	assert.Contains(t, err.Error(), validConnectPeerID())
	assert.NotContains(t, err.Error(), "timed out by command context deadline")
}

func TestConnectToPeer_CleansUpAfterConnectFailure(t *testing.T) {
	testCases := []struct {
		name       string
		connectErr error
	}{
		{name: "timeout", connectErr: context.DeadlineExceeded},
		{name: "canceled", connectErr: context.Canceled},
		{name: "other", connectErr: errors.New("dial refused")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cleanedUp := false
			restoreConnectDeps(t, connectDeps{
				config: connectTestP2PConfig(time.Second),
				connect: func(context.Context, peer.AddrInfo) error {
					return tc.connectErr
				},
				cleanup: func() {
					cleanedUp = true
				},
			})

			_, cleanup, err := connectToPeer(context.Background(), &bootstrap.Result{}, validConnectMultiaddr())
			require.Error(t, err)
			assert.Nil(t, cleanup)
			assert.True(t, cleanedUp, "cleanup should run on connect failure")
			assert.Contains(t, err.Error(), validConnectPeerID())
		})
	}
}

func restoreConnectDeps(t *testing.T, deps connectDeps) {
	t.Helper()
	original := loadConnectDeps
	loadConnectDeps = func(*bootstrap.Result) (connectDeps, error) {
		return deps, nil
	}
	t.Cleanup(func() { loadConnectDeps = original })
}

func connectTestP2PConfig(timeout time.Duration) *config.P2PConfig {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.HandshakeTimeout = timeout
	return &cfg.P2P
}

func validConnectMultiaddr() string {
	return "/ip4/127.0.0.1/tcp/9000/p2p/" + validConnectPeerID()
}

func validConnectPeerID() string {
	return "12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN"
}

func TestDisconnectCmd_WritesToCommandWriter(t *testing.T) {
	original := disconnectFromPeer
	disconnectFromPeer = func(_ *bootstrap.Result, target string) (string, func(), error) {
		require.Equal(t, "12D3KooWDisconnect", target)
		return "12D3KooWDisconnect", func() {}, nil
	}
	t.Cleanup(func() { disconnectFromPeer = original })

	cmd := newDisconnectCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeP2PCmd(t, cmd, "12D3KooWDisconnect")
	require.NoError(t, err)
	assert.Contains(t, out, "Disconnected from peer 12D3KooWDisconnect")
}

func TestProvenancePushCmd_WritesToCommandWriter(t *testing.T) {
	original := provenancePostJSON
	provenancePostJSON = func(addr, path string, body any, out any) error {
		require.Equal(t, "http://127.0.0.1:7777", addr)
		require.Equal(t, "/api/p2p/provenance/push", path)
		payload, ok := body.(map[string]string)
		require.True(t, ok)
		require.Equal(t, "did:lango:peer1", payload["peerDid"])
		require.Equal(t, "session-key-1", payload["sessionKey"])
		require.Equal(t, "full", payload["redaction"])
		if outMap, ok := out.(*map[string]any); ok {
			*outMap = map[string]any{"ok": true}
		}
		return nil
	}
	t.Cleanup(func() { provenancePostJSON = original })

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cmd := newProvenancePushCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeP2PCmd(t, cmd, "did:lango:peer1", "session-key-1", "--redaction", "full", "--addr", "http://127.0.0.1:7777")
	require.NoError(t, err)
	assert.Contains(t, out, "Pushed provenance bundle to did:lango:peer1 (redaction=full)")
}

func TestProvenanceFetchCmd_WritesToCommandWriter(t *testing.T) {
	original := provenancePostJSON
	provenancePostJSON = func(addr, path string, body any, out any) error {
		require.Equal(t, "http://127.0.0.1:8888", addr)
		require.Equal(t, "/api/p2p/provenance/fetch", path)
		if outMap, ok := out.(*map[string]any); ok {
			*outMap = map[string]any{"redaction": "content"}
		}
		return nil
	}
	t.Cleanup(func() { provenancePostJSON = original })

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cmd := newProvenanceFetchCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeP2PCmd(t, cmd, "did:lango:peer2", "session-key-2", "--addr", "http://127.0.0.1:8888")
	require.NoError(t, err)
	assert.Contains(t, out, "Fetched provenance bundle from did:lango:peer2 (redaction=content)")
}

func TestPricingCmd_WritesTextToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Pricing.Enabled = true
	cfg.P2P.Pricing.PerQuery = "0.05"
	cfg.P2P.Pricing.ToolPrices = map[string]string{
		"knowledge_search": "0.15",
	}

	cmd := newPricingCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "P2P Pricing Configuration")
	assert.Contains(t, out, "Enabled:     true")
	assert.Contains(t, out, "Per Query:   0.05 USDC")
	assert.Contains(t, out, "knowledge_search")
}

func TestPricingCmd_WritesToolViewToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Pricing.PerQuery = "0.05"
	cfg.P2P.Pricing.ToolPrices = map[string]string{
		"knowledge_search": "0.15",
	}

	cmd := newPricingCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--tool", "knowledge_search")
	require.NoError(t, err)
	assert.Contains(t, out, "Tool:     knowledge_search")
	assert.Contains(t, out, "Price:    0.15 USDC")
}

func TestPricingCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Pricing.Enabled = false
	cfg.P2P.Pricing.PerQuery = "0.07"
	cfg.P2P.Pricing.ToolPrices = map[string]string{
		"knowledge_search": "0.15",
	}

	cmd := newPricingCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, false, decoded["enabled"])
	assert.Equal(t, "0.07", decoded["perQuery"])
	assert.Equal(t, "USDC", decoded["currency"])
}

func TestZKPStatusCmd_WritesTextToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.ZKHandshake = true
	cfg.P2P.ZKAttestation = false
	cfg.P2P.ZKP.ProvingScheme = "plonk"
	cfg.P2P.ZKP.SRSMode = "local"
	cfg.P2P.ZKP.SRSPath = "/tmp/test.srs"
	cfg.P2P.ZKP.ProofCacheDir = "/tmp/proofs"
	cfg.P2P.ZKP.MaxCredentialAge = "24h"

	cmd := newZKPStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "ZKP Configuration")
	assert.Contains(t, out, "Proving Scheme:     plonk")
	assert.Contains(t, out, "SRS Path:           /tmp/test.srs")
}

func TestZKPStatusCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.ZKHandshake = false
	cfg.P2P.ZKAttestation = true
	cfg.P2P.ZKP.ProvingScheme = "groth16"
	cfg.P2P.ZKP.SRSMode = "embedded"

	cmd := newZKPStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, false, decoded["zkHandshake"])
	assert.Equal(t, true, decoded["zkAttestation"])
	assert.Equal(t, "groth16", decoded["provingScheme"])
}

func TestZKPCircuitsCmd_WritesTextToCommandWriter(t *testing.T) {
	cmd := newZKPCircuitsCmd()

	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "CIRCUIT")
	assert.Contains(t, out, "identity")
	assert.Contains(t, out, "reputation")
}

func TestZKPCircuitsCmd_WritesJSONToCommandWriter(t *testing.T) {
	cmd := newZKPCircuitsCmd()

	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.NotEmpty(t, decoded)
	assert.Equal(t, "identity", decoded[0]["id"])
}

func TestFirewallCmd_HasSubcommands(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "firewall" {
			firewallSubs := make(map[string]bool)
			for _, fsub := range sub.Commands() {
				firewallSubs[fsub.Use] = true
			}
			assert.True(t, firewallSubs["list"], "firewall should have list subcommand")
			assert.True(t, firewallSubs["add"], "firewall should have add subcommand")
			assert.True(t, firewallSubs["remove <peer-did>"], "firewall should have remove subcommand")
			return
		}
	}
	t.Fatal("firewall subcommand not found")
}

func TestSessionCmd_HasSubcommands(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "session" {
			sessionSubs := make(map[string]bool)
			for _, ssub := range sub.Commands() {
				sessionSubs[ssub.Use] = true
			}
			assert.True(t, sessionSubs["list"], "session should have list subcommand")
			assert.True(t, sessionSubs["revoke"], "session should have revoke subcommand")
			assert.True(t, sessionSubs["revoke-all"], "session should have revoke-all subcommand")
			return
		}
	}
	t.Fatal("session subcommand not found")
}

func TestSandboxCmd_HasSubcommands(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "sandbox" {
			sandboxSubs := make(map[string]bool)
			for _, ssub := range sub.Commands() {
				sandboxSubs[ssub.Use] = true
			}
			assert.True(t, sandboxSubs["status"], "sandbox should have status subcommand")
			assert.True(t, sandboxSubs["test"], "sandbox should have test subcommand")
			assert.True(t, sandboxSubs["cleanup"], "sandbox should have cleanup subcommand")
			return
		}
	}
	t.Fatal("sandbox subcommand not found")
}
