package protocol

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRemoteAgent_FieldPopulation(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	peerID, err := peer.Decode("12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN")
	require.NoError(t, err)

	cfg := RemoteAgentConfig{
		Name:         "test-agent",
		DID:          "did:key:z6MkpTHR8VNs9bN38RNsB",
		PeerID:       peerID,
		SessionToken: "tok-abc",
		Host:         nil,
		Capabilities: []string{"search", "translate"},
		Logger:       sugar,
	}

	agent := NewRemoteAgent(cfg)
	require.NotNil(t, agent)

	assert.Equal(t, "test-agent", agent.name)
	assert.Equal(t, "did:key:z6MkpTHR8VNs9bN38RNsB", agent.did)
	assert.Equal(t, peerID, agent.peerID)
	assert.Equal(t, "tok-abc", agent.token)
	assert.Nil(t, agent.host)
	assert.Equal(t, []string{"search", "translate"}, agent.capabilities)
	assert.Nil(t, agent.attestVerify)
	assert.Equal(t, sugar, agent.logger)
}

func TestNewRemoteAgent_WithAttestVerifier(t *testing.T) {
	t.Parallel()

	verifier := func(_ context.Context, _ *AttestationData) (bool, error) {
		return true, nil
	}

	cfg := RemoteAgentConfig{
		Name:           "verified-agent",
		DID:            "did:key:z6Mkverified",
		AttestVerifier: verifier,
	}

	agent := NewRemoteAgent(cfg)
	require.NotNil(t, agent.attestVerify)

	ok, err := agent.attestVerify(context.Background(), &AttestationData{})
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestP2PRemoteAgent_Accessors(t *testing.T) {
	t.Parallel()

	peerID, err := peer.Decode("12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN")
	require.NoError(t, err)

	tests := []struct {
		give     RemoteAgentConfig
		wantName string
		wantDID  string
		wantPeer peer.ID
		wantCaps []string
	}{
		{
			give: RemoteAgentConfig{
				Name:         "agent-alpha",
				DID:          "did:key:alpha",
				PeerID:       peerID,
				Capabilities: []string{"cap1", "cap2"},
			},
			wantName: "agent-alpha",
			wantDID:  "did:key:alpha",
			wantPeer: peerID,
			wantCaps: []string{"cap1", "cap2"},
		},
		{
			give: RemoteAgentConfig{
				Name:   "agent-empty-caps",
				DID:    "did:key:empty",
				PeerID: peerID,
			},
			wantName: "agent-empty-caps",
			wantDID:  "did:key:empty",
			wantPeer: peerID,
			wantCaps: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			t.Parallel()

			agent := NewRemoteAgent(tt.give)

			assert.Equal(t, tt.wantName, agent.Name())
			assert.Equal(t, tt.wantDID, agent.DID())
			assert.Equal(t, tt.wantPeer, agent.PeerID())
			assert.Equal(t, tt.wantCaps, agent.Capabilities())
		})
	}
}

func TestP2PRemoteAgent_SetAttestVerifier(t *testing.T) {
	t.Parallel()

	agent := NewRemoteAgent(RemoteAgentConfig{
		Name: "agent-no-verifier",
		DID:  "did:key:z6MkNoVerifier",
	})

	// Initially nil.
	assert.Nil(t, agent.attestVerify)

	// Set verifier.
	called := false
	agent.SetAttestVerifier(func(_ context.Context, _ *AttestationData) (bool, error) {
		called = true
		return false, nil
	})

	require.NotNil(t, agent.attestVerify)

	ok, err := agent.attestVerify(context.Background(), nil)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.True(t, called)
}

func TestP2PRemoteAgent_ZeroValueConfig(t *testing.T) {
	t.Parallel()

	agent := NewRemoteAgent(RemoteAgentConfig{})
	require.NotNil(t, agent)

	assert.Equal(t, "", agent.Name())
	assert.Equal(t, "", agent.DID())
	assert.Equal(t, peer.ID(""), agent.PeerID())
	assert.Nil(t, agent.Capabilities())
}
