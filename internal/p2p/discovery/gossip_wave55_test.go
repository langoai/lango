package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/require"
)

func TestPeerIDFromStringWave55_ParsesValidAndRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	valid := "QmYwAPJzv5CZsnAzt8auVZRnGPZUEu4DqQx9fBBRgKjZ8M"
	peerID, err := PeerIDFromString(valid)
	require.NoError(t, err)
	require.Equal(t, valid, peerID.String())

	_, err = PeerIDFromString("not-a-peer-id")
	require.Error(t, err)
}

func TestGossipServiceHandleMessageWave55_StoresUnsignedNewerCardsOnly(t *testing.T) {
	t.Parallel()

	gs := newTestGossipServiceFields()
	older := GossipCard{
		DID:          "did:lango:wave55",
		Name:         "older",
		Capabilities: []string{"search"},
		Timestamp:    time.Unix(100, 0).UTC(),
	}
	newer := older
	newer.Name = "newer"
	newer.Timestamp = time.Unix(200, 0).UTC()

	gs.handleMessage(wave55PubSubMessage(t, newer))
	require.Equal(t, "newer", gs.FindByDID("did:lango:wave55").Name)

	gs.handleMessage(wave55PubSubMessage(t, older))
	require.Equal(t, "newer", gs.FindByDID("did:lango:wave55").Name)
	require.Len(t, gs.FindByCapability("search"), 1)
	require.Len(t, gs.KnownPeers(), 1)
}

func TestGossipServiceHandleMessageWave55_ValidationRejectsVerifierErrorsAndMalformedSignedBundles(t *testing.T) {
	t.Parallel()

	t.Run("credential verifier error discards card", func(t *testing.T) {
		t.Parallel()

		gs := newTestGossipServiceFields()
		gs.verifier = func(*ZKCredential) (bool, error) {
			return false, errors.New("proof backend unavailable")
		}

		gs.handleMessage(wave55PubSubMessage(t, GossipCard{
			DID:       "did:lango:bad-proof",
			Name:      "bad-proof",
			Timestamp: time.Unix(300, 0).UTC(),
			ZKCredentials: []ZKCredential{{
				CapabilityID: "search",
				IssuedAt:     time.Now(),
				ExpiresAt:    time.Now().Add(time.Hour),
			}},
		}))

		require.Nil(t, gs.FindByDID("did:lango:bad-proof"))
	})

	t.Run("signed card with malformed bundle is discarded", func(t *testing.T) {
		t.Parallel()

		gs := newTestGossipServiceFields()
		gs.handleMessage(wave55PubSubMessage(t, GossipCard{
			DID:       "did:lango:bad-bundle",
			Name:      "bad-bundle",
			Bundle:    json.RawMessage(`{"signing_key":{"public_key":null}}`),
			Signature: []byte("signature"),
			Timestamp: time.Unix(400, 0).UTC(),
		}))

		require.Nil(t, gs.FindByDID("did:lango:bad-bundle"))
	})
}

func TestGossipServicePublishLoopWave55_CanceledContextReturnsWithoutHostOrTopic(t *testing.T) {
	t.Parallel()

	gs := newTestGossipServiceFields()
	gs.localCard = nil
	gs.interval = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	gs.publishLoop(ctx)

	require.Empty(t, gs.KnownPeers())
}

func wave55PubSubMessage(t *testing.T, card GossipCard) *pubsub.Message {
	t.Helper()

	data, err := json.Marshal(card)
	require.NoError(t, err)
	return &pubsub.Message{Message: &pb.Message{Data: data}}
}
