package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/p2p/identity"
)

type wave10CardSigner struct {
	alg     string
	sig     []byte
	err     error
	payload []byte
}

func (s *wave10CardSigner) Sign(_ context.Context, payload []byte) ([]byte, error) {
	s.payload = append([]byte(nil), payload...)
	if s.err != nil {
		return nil, s.err
	}
	return append([]byte(nil), s.sig...), nil
}

func (s *wave10CardSigner) Algorithm() string {
	return s.alg
}

type wave10PQSigner struct {
	alg     string
	pub     []byte
	sig     []byte
	err     error
	payload []byte
}

func (s *wave10PQSigner) SignPQ(_ context.Context, payload []byte) ([]byte, error) {
	s.payload = append([]byte(nil), payload...)
	if s.err != nil {
		return nil, s.err
	}
	return append([]byte(nil), s.sig...), nil
}

func (s *wave10PQSigner) PQAlgorithm() string {
	return s.alg
}

func (s *wave10PQSigner) PQPublicKey() []byte {
	return append([]byte(nil), s.pub...)
}

func TestCanonicalCardPayloadWave10_IgnoresSignaturesButKeepsAlgorithms(t *testing.T) {
	t.Parallel()

	card := &GossipCard{
		DID:                  "did:lango:canonical",
		Name:                 "canonical-agent",
		PeerID:               "peer-a",
		Timestamp:            time.Unix(1700000000, 0).UTC(),
		Signature:            []byte("sig-a"),
		SignatureAlgorithm:   "ed25519",
		PQSignerPublicKey:    []byte("pq-public"),
		PQSignature:          []byte("pq-sig-a"),
		PQSignatureAlgorithm: "mldsa65",
	}

	first, err := CanonicalCardPayload(card)
	require.NoError(t, err)

	card.Signature = []byte("sig-b")
	card.PQSignature = []byte("pq-sig-b")
	second, err := CanonicalCardPayload(card)
	require.NoError(t, err)
	assert.Equal(t, first, second)

	card.SignatureAlgorithm = "other-algorithm"
	third, err := CanonicalCardPayload(card)
	require.NoError(t, err)
	assert.NotEqual(t, first, third)
	assert.NotContains(t, string(first), "sig-a")
	assert.Contains(t, string(first), "ed25519")
	assert.Contains(t, string(first), "mldsa65")
}

func TestVerifyCardSignatureWave10_UsesCanonicalPayloadAndBundleKey(t *testing.T) {
	t.Parallel()

	bundle := wave10IdentityBundle()
	did, err := identity.ComputeDIDv2(bundle)
	require.NoError(t, err)
	bundleJSON, err := json.Marshal(bundle)
	require.NoError(t, err)

	card := &GossipCard{
		DID:                did,
		Name:               "signed-agent",
		Bundle:             bundleJSON,
		Signature:          []byte("signature"),
		SignatureAlgorithm: "ed25519",
		Timestamp:          time.Unix(1700000000, 0).UTC(),
	}
	wantPayload, err := CanonicalCardPayload(card)
	require.NoError(t, err)

	var gotPubkey, gotPayload, gotSig []byte
	err = VerifyCardSignature(card, func(pubkey, message, sig []byte) error {
		gotPubkey = append([]byte(nil), pubkey...)
		gotPayload = append([]byte(nil), message...)
		gotSig = append([]byte(nil), sig...)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, bundle.SigningKey.PublicKey, gotPubkey)
	assert.Equal(t, wantPayload, gotPayload)
	assert.Equal(t, []byte("signature"), gotSig)
}

func TestVerifyCardSignatureWave10_RejectsInvalidSignedCards(t *testing.T) {
	t.Parallel()

	bundle := wave10IdentityBundle()
	did, err := identity.ComputeDIDv2(bundle)
	require.NoError(t, err)
	bundleJSON, err := json.Marshal(bundle)
	require.NoError(t, err)

	tests := []struct {
		name       string
		card       *GossipCard
		verifyFunc func(pubkey, message, sig []byte) error
		want       string
	}{
		{
			name: "bundle has no signing key",
			card: &GossipCard{
				DID:       did,
				Bundle:    []byte(`{"signing_key":{"public_key":null}}`),
				Signature: []byte("signature"),
			},
			want: "no valid signing key",
		},
		{
			name: "classical verifier rejects signature",
			card: &GossipCard{
				DID:       did,
				Bundle:    bundleJSON,
				Signature: []byte("signature"),
			},
			verifyFunc: func(pubkey, message, sig []byte) error {
				return errors.New("bad signature")
			},
			want: "bad signature",
		},
		{
			name: "DID does not match bundle",
			card: &GossipCard{
				DID:       "did:lango:v2:0000000000000000000000000000000000000000",
				Bundle:    bundleJSON,
				Signature: []byte("signature"),
			},
			want: "does not match bundle",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := VerifyCardSignature(tt.card, tt.verifyFunc)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestVerifyCardSignatureWave10_AllowsUnsignedLegacyCard(t *testing.T) {
	t.Parallel()

	err := VerifyCardSignature(&GossipCard{DID: "did:lango:legacy"}, func(pubkey, message, sig []byte) error {
		t.Fatal("verifier should not be called for unsigned cards")
		return nil
	})
	require.NoError(t, err)
}

func TestGossipServiceHandleMessageWave10_FiltersInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(*GossipService)
		data  []byte
	}{
		{
			name: "invalid JSON",
			data: []byte("{"),
		},
		{
			name: "missing DID",
			data: wave10MarshalCard(t, GossipCard{Name: "missing-did"}),
		},
		{
			name: "revoked DID",
			setup: func(gs *GossipService) {
				gs.revokedDIDs["did:lango:revoked"] = time.Now()
			},
			data: wave10MarshalCard(t, GossipCard{
				DID:       "did:lango:revoked",
				Timestamp: time.Unix(1700000000, 0).UTC(),
			}),
		},
		{
			name: "invalid signature",
			setup: func(gs *GossipService) {
				gs.classicalVerify = func(pubkey, message, sig []byte) error {
					return errors.New("signature rejected")
				}
			},
			data: wave10SignedCardJSON(t),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gs := newTestGossipServiceFields()
			if tt.setup != nil {
				tt.setup(gs)
			}

			gs.handleMessage(&pubsub.Message{Message: &pb.Message{Data: tt.data}})
			assert.Empty(t, gs.KnownPeers())
		})
	}
}

func TestGossipServiceHandleMessageWave10_CredentialAndTimestampBranches(t *testing.T) {
	t.Parallel()

	now := time.Now()
	verifierCalls := 0
	gs := newTestGossipServiceFields()
	gs.maxCredentialAge = time.Hour
	gs.verifier = func(cred *ZKCredential) (bool, error) {
		verifierCalls++
		return cred.CapabilityID != "invalid", nil
	}

	staleButAccepted := GossipCard{
		DID:       "did:lango:cred",
		Name:      "stale",
		Timestamp: now,
		ZKCredentials: []ZKCredential{
			{CapabilityID: "expired", IssuedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(-time.Minute)},
			{CapabilityID: "stale", IssuedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(time.Hour)},
		},
	}
	gs.handleMessage(&pubsub.Message{Message: &pb.Message{Data: wave10MarshalCard(t, staleButAccepted)}})
	require.NotNil(t, gs.FindByDID("did:lango:cred"))
	assert.Equal(t, 0, verifierCalls)

	older := staleButAccepted
	older.Name = "older"
	older.Timestamp = now.Add(-time.Minute)
	gs.handleMessage(&pubsub.Message{Message: &pb.Message{Data: wave10MarshalCard(t, older)}})
	assert.Equal(t, "stale", gs.FindByDID("did:lango:cred").Name)

	newer := staleButAccepted
	newer.Name = "newer"
	newer.Timestamp = now.Add(time.Minute)
	gs.handleMessage(&pubsub.Message{Message: &pb.Message{Data: wave10MarshalCard(t, newer)}})
	assert.Equal(t, "newer", gs.FindByDID("did:lango:cred").Name)

	invalid := GossipCard{
		DID:       "did:lango:invalid-cred",
		Name:      "invalid",
		Timestamp: now,
		ZKCredentials: []ZKCredential{
			{CapabilityID: "invalid", IssuedAt: now, ExpiresAt: now.Add(time.Hour)},
		},
	}
	gs.handleMessage(&pubsub.Message{Message: &pb.Message{Data: wave10MarshalCard(t, invalid)}})
	assert.Nil(t, gs.FindByDID("did:lango:invalid-cred"))
	assert.Equal(t, 1, verifierCalls)
}

func TestGossipServiceSignCardWave10_Branches(t *testing.T) {
	t.Parallel()

	t.Run("nil signer leaves card unsigned", func(t *testing.T) {
		t.Parallel()

		card := &GossipCard{DID: "did:lango:unsigned"}
		gs := newTestGossipServiceFields()
		gs.signCard(context.Background(), card)
		assert.Empty(t, card.Signature)
		assert.Empty(t, card.SignatureAlgorithm)
	})

	t.Run("classical and PQ signatures use the same canonical payload", func(t *testing.T) {
		t.Parallel()

		card := &GossipCard{DID: "did:lango:signed", Timestamp: time.Unix(1700000000, 0).UTC()}
		signer := &wave10CardSigner{alg: "ed25519", sig: []byte("classical-signature")}
		pqSigner := &wave10PQSigner{alg: "mldsa65", pub: []byte("pq-public"), sig: []byte("pq-signature")}
		gs := newTestGossipServiceFields()
		gs.cardSigner = signer
		gs.pqSigner = pqSigner

		gs.signCard(context.Background(), card)

		assert.Equal(t, "ed25519", card.SignatureAlgorithm)
		assert.Equal(t, []byte("classical-signature"), card.Signature)
		assert.Equal(t, "mldsa65", card.PQSignatureAlgorithm)
		assert.Equal(t, []byte("pq-public"), card.PQSignerPublicKey)
		assert.Equal(t, []byte("pq-signature"), card.PQSignature)
		assert.NotEmpty(t, signer.payload)
		assert.Equal(t, signer.payload, pqSigner.payload)
		assert.False(t, bytes.Contains(signer.payload, []byte("classical-signature")))
		assert.False(t, bytes.Contains(signer.payload, []byte("pq-signature")))
	})

	t.Run("classical signing error skips signatures", func(t *testing.T) {
		t.Parallel()

		card := &GossipCard{DID: "did:lango:sign-error"}
		gs := newTestGossipServiceFields()
		gs.cardSigner = &wave10CardSigner{alg: "ed25519", err: errors.New("sign failed")}

		gs.signCard(context.Background(), card)

		assert.Equal(t, "ed25519", card.SignatureAlgorithm)
		assert.Empty(t, card.Signature)
	})
}

func wave10IdentityBundle() *identity.IdentityBundle {
	return &identity.IdentityBundle{
		Version: 1,
		SigningKey: identity.PublicKeyEntry{
			Algorithm: "ed25519",
			PublicKey: bytes.Repeat([]byte{0x11}, 32),
		},
		SettlementKey: identity.PublicKeyEntry{
			Algorithm: "secp256k1",
			PublicKey: bytes.Repeat([]byte{0x22}, 33),
		},
		LegacyDID: "did:lango:legacy",
	}
}

func wave10SignedCardJSON(t *testing.T) []byte {
	t.Helper()

	bundle := wave10IdentityBundle()
	did, err := identity.ComputeDIDv2(bundle)
	require.NoError(t, err)
	bundleJSON, err := json.Marshal(bundle)
	require.NoError(t, err)
	return wave10MarshalCard(t, GossipCard{
		DID:       did,
		Bundle:    bundleJSON,
		Signature: []byte("signature"),
		Timestamp: time.Unix(1700000000, 0).UTC(),
	})
}

func wave10MarshalCard(t *testing.T, card GossipCard) []byte {
	t.Helper()

	data, err := json.Marshal(card)
	require.NoError(t, err)
	return data
}
