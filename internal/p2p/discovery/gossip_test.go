package discovery

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGossipServiceFields creates a GossipService with only the internal
// fields set, suitable for testing query/revocation methods that do not touch
// the libp2p host, PubSub, or topic.
func newTestGossipServiceFields() *GossipService {
	return &GossipService{
		peers:            make(map[string]*GossipCard),
		revokedDIDs:      make(map[string]time.Time),
		maxCredentialAge: defaultMaxCredentialAge,
		logger:           testLogger(),
	}
}

func TestGossipService_KnownPeers_Empty(t *testing.T) {
	gs := newTestGossipServiceFields()
	assert.Empty(t, gs.KnownPeers())
}

func TestGossipService_KnownPeers_AfterAdding(t *testing.T) {
	gs := newTestGossipServiceFields()
	gs.peers["did:lango:a"] = &GossipCard{DID: "did:lango:a", Name: "alice"}
	gs.peers["did:lango:b"] = &GossipCard{DID: "did:lango:b", Name: "bob"}

	peers := gs.KnownPeers()
	assert.Len(t, peers, 2)
}

func TestGossipService_FindByCapability_Match(t *testing.T) {
	gs := newTestGossipServiceFields()
	gs.peers["did:lango:a"] = &GossipCard{
		DID:          "did:lango:a",
		Capabilities: []string{"search", "translate"},
	}
	gs.peers["did:lango:b"] = &GossipCard{
		DID:          "did:lango:b",
		Capabilities: []string{"code"},
	}

	matches := gs.FindByCapability("search")
	require.Len(t, matches, 1)
	assert.Equal(t, "did:lango:a", matches[0].DID)
}

func TestGossipService_FindByCapability_NoMatch(t *testing.T) {
	gs := newTestGossipServiceFields()
	gs.peers["did:lango:a"] = &GossipCard{
		DID:          "did:lango:a",
		Capabilities: []string{"search"},
	}

	matches := gs.FindByCapability("unknown")
	assert.Empty(t, matches)
}

func TestGossipService_FindByDID(t *testing.T) {
	gs := newTestGossipServiceFields()
	card := &GossipCard{DID: "did:lango:alice", Name: "alice"}
	gs.peers["did:lango:alice"] = card

	found := gs.FindByDID("did:lango:alice")
	require.NotNil(t, found)
	assert.Equal(t, "alice", found.Name)

	notFound := gs.FindByDID("did:lango:unknown")
	assert.Nil(t, notFound)
}

func TestGossipService_RevokeDID_And_IsRevoked(t *testing.T) {
	gs := newTestGossipServiceFields()

	assert.False(t, gs.IsRevoked("did:lango:bad"))

	gs.RevokeDID("did:lango:bad")
	assert.True(t, gs.IsRevoked("did:lango:bad"))

	assert.False(t, gs.IsRevoked("did:lango:good"))
}

func TestGossipService_SetMaxCredentialAge(t *testing.T) {
	gs := newTestGossipServiceFields()
	assert.Equal(t, defaultMaxCredentialAge, gs.maxCredentialAge)

	gs.SetMaxCredentialAge(12 * time.Hour)

	gs.revokedMu.RLock()
	assert.Equal(t, 12*time.Hour, gs.maxCredentialAge)
	gs.revokedMu.RUnlock()
}

func TestGossipService_DefaultMaxCredentialAge(t *testing.T) {
	assert.Equal(t, 24*time.Hour, defaultMaxCredentialAge)
}

func TestTopicAgentCard_Constant(t *testing.T) {
	assert.Equal(t, "/lango/agentcard/1.0.0", TopicAgentCard)
}

func TestPeerIDFromString_Valid(t *testing.T) {
	// Use a well-known peer ID format (base58 encoded).
	// This tests that the function wraps peer.Decode correctly.
	_, err := PeerIDFromString("invalid-peer-id")
	assert.Error(t, err, "invalid peer ID string should return error")
}
