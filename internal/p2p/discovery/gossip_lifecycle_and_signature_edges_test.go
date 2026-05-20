package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGossipServiceLifecycleUsesConfiguredFieldsWithoutListening(t *testing.T) {
	h, err := libp2p.New(libp2p.NoListenAddrs)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, h.Close())
	})

	gs, err := NewGossipService(GossipConfig{
		Host:      h,
		Interval:  time.Hour,
		Verifier:  func(*ZKCredential) (bool, error) { return true, nil },
		Logger:    testLogger(),
		LocalCard: nil,
	})
	require.NoError(t, err)
	require.NotNil(t, gs.topic)
	require.NotNil(t, gs.sub)
	assert.Same(t, h, gs.host)
	assert.Empty(t, gs.KnownPeers())
	assert.Equal(t, defaultMaxCredentialAge, gs.maxCredentialAge)

	var wg sync.WaitGroup
	gs.Start(&wg)
	require.NotNil(t, gs.cancel)
	gs.Stop()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("gossip service goroutines did not stop after Stop")
	}
}

func TestVerifyCardSignatureReturnsCanonicalPayloadErrors(t *testing.T) {
	t.Parallel()

	err := VerifyCardSignature(&GossipCard{
		DID:       "did:lango:bad-json",
		Bundle:    json.RawMessage(`{`),
		Signature: []byte("signature"),
	}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "canonical card payload")
}

func TestGossipServiceSignCardHandlesCanonicalAndPQErrors(t *testing.T) {
	t.Parallel()

	t.Run("canonical payload error skips signing", func(t *testing.T) {
		t.Parallel()

		signer := &canonicalCardPayloadIgnoresSignaturesButKeepsAlgorithmsCardSigner{
			alg: "ed25519",
			sig: []byte("signature"),
		}
		gs := newTestGossipServiceFields()
		gs.cardSigner = signer
		card := &GossipCard{
			DID:    "did:lango:canonical-error",
			Bundle: json.RawMessage(`{`),
		}

		gs.signCard(context.Background(), card)

		assert.Equal(t, "ed25519", card.SignatureAlgorithm)
		assert.Empty(t, signer.payload)
		assert.Empty(t, card.Signature)
	})

	t.Run("PQ signing error keeps classical signature and omits PQ signature", func(t *testing.T) {
		t.Parallel()

		signer := &canonicalCardPayloadIgnoresSignaturesButKeepsAlgorithmsCardSigner{
			alg: "ed25519",
			sig: []byte("classical-signature"),
		}
		pqSigner := &canonicalCardPayloadIgnoresSignaturesButKeepsAlgorithmsPQSigner{
			alg: "mldsa65",
			pub: []byte("pq-public"),
			err: errors.New("pq signer offline"),
		}
		gs := newTestGossipServiceFields()
		gs.cardSigner = signer
		gs.pqSigner = pqSigner
		card := &GossipCard{
			DID:       "did:lango:pq-error",
			Timestamp: time.Unix(1700000100, 0).UTC(),
		}

		gs.signCard(context.Background(), card)

		assert.Equal(t, []byte("classical-signature"), card.Signature)
		assert.Equal(t, "mldsa65", card.PQSignatureAlgorithm)
		assert.Equal(t, []byte("pq-public"), card.PQSignerPublicKey)
		assert.Empty(t, card.PQSignature)
		assert.Equal(t, signer.payload, pqSigner.payload)
	})
}
