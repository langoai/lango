package handshake

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/security"
)

// mockEd25519Signer implements Signer for Ed25519 (framework testing only).
type mockEd25519Signer struct {
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

func (m *mockEd25519Signer) SignMessage(_ context.Context, message []byte) ([]byte, error) {
	return ed25519.Sign(m.priv, message), nil
}

func (m *mockEd25519Signer) PublicKey(_ context.Context) ([]byte, error) {
	return []byte(m.pub), nil
}

func (m *mockEd25519Signer) Algorithm() string { return security.AlgorithmEd25519 }

func (m *mockEd25519Signer) DID(_ context.Context) (string, error) {
	return "did:lango:v2:" + hex.EncodeToString(m.pub[:20]), nil
}

func ed25519GenerateKey() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func ed25519Sign(priv ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(priv, message)
}

// mockSigner implements the Signer interface for testing.
type mockSigner struct {
	privKeyBytes []byte
}

func (m *mockSigner) SignMessage(_ context.Context, message []byte) ([]byte, error) {
	key, err := ethcrypto.ToECDSA(m.privKeyBytes)
	if err != nil {
		return nil, err
	}
	hash := ethcrypto.Keccak256(message)
	return ethcrypto.Sign(hash, key)
}

func (m *mockSigner) PublicKey(_ context.Context) ([]byte, error) {
	key, err := ethcrypto.ToECDSA(m.privKeyBytes)
	if err != nil {
		return nil, err
	}
	return ethcrypto.CompressPubkey(&key.PublicKey), nil
}

func (m *mockSigner) Algorithm() string { return security.AlgorithmSecp256k1Keccak256 }

func (m *mockSigner) DID(_ context.Context) (string, error) {
	pub, err := m.PublicKey(context.Background())
	if err != nil {
		return "", err
	}
	return "did:lango:" + fmt.Sprintf("%x", pub), nil
}

func newTestHandshaker(t *testing.T, s *mockSigner) *Handshaker {
	t.Helper()
	sessions, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	return NewHandshaker(Config{
		Signer:   s,
		Sessions: sessions,
		Timeout:  30 * time.Second,
		Logger:   zap.NewNop().Sugar(),
	})
}

func TestVerifyResponse_ValidSignature(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := ethcrypto.FromECDSA(privKey)

	w := &mockSigner{privKeyBytes: privBytes}
	h := newTestHandshaker(t, w)

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	sig, err := w.SignMessage(context.Background(), nonce)
	require.NoError(t, err)

	pubkey, err := w.PublicKey(context.Background())
	require.NoError(t, err)

	resp := &ChallengeResponse{
		Nonce:     nonce,
		Signature: sig,
		PublicKey: pubkey,
		DID:       "did:lango:test",
	}

	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.NoError(t, err)
}

func TestVerifyResponse_InvalidSignature(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := ethcrypto.FromECDSA(privKey)

	w := &mockSigner{privKeyBytes: privBytes}
	h := newTestHandshaker(t, w)

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")

	// Sign with one key but claim a different public key.
	sig, err := w.SignMessage(context.Background(), nonce)
	require.NoError(t, err)

	otherKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	otherPubkey := ethcrypto.CompressPubkey(&otherKey.PublicKey)

	resp := &ChallengeResponse{
		Nonce:     nonce,
		Signature: sig,
		PublicKey: otherPubkey,
		DID:       "did:lango:test",
	}

	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public key mismatch")
}

func TestVerifyResponse_WrongSignatureLength(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := ethcrypto.FromECDSA(privKey)

	w := &mockSigner{privKeyBytes: privBytes}
	h := newTestHandshaker(t, w)

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	pubkey, err := w.PublicKey(context.Background())
	require.NoError(t, err)

	resp := &ChallengeResponse{
		Nonce:     nonce,
		Signature: []byte("too-short"),
		PublicKey: pubkey,
		DID:       "did:lango:test",
	}

	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature length")
}

func TestVerifyResponse_NonceMismatch(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := ethcrypto.FromECDSA(privKey)

	w := &mockSigner{privKeyBytes: privBytes}
	h := newTestHandshaker(t, w)

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	wrongNonce := []byte("wrong-nonce-does-not-match!!!!!!!")

	sig, err := w.SignMessage(context.Background(), nonce)
	require.NoError(t, err)
	pubkey, err := w.PublicKey(context.Background())
	require.NoError(t, err)

	resp := &ChallengeResponse{
		Nonce:     wrongNonce,
		Signature: sig,
		PublicKey: pubkey,
		DID:       "did:lango:test",
	}

	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonce mismatch")
}

func TestVerifyResponse_NoProofOrSignature(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := ethcrypto.FromECDSA(privKey)

	w := &mockSigner{privKeyBytes: privBytes}
	h := newTestHandshaker(t, w)

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	pubkey, err := w.PublicKey(context.Background())
	require.NoError(t, err)

	resp := &ChallengeResponse{
		Nonce:     nonce,
		PublicKey: pubkey,
		DID:       "did:lango:test",
	}

	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no proof or signature")
}

func TestVerifyResponse_CorruptedSignature(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := ethcrypto.FromECDSA(privKey)

	w := &mockSigner{privKeyBytes: privBytes}
	h := newTestHandshaker(t, w)

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	sig, err := w.SignMessage(context.Background(), nonce)
	require.NoError(t, err)
	pubkey, err := w.PublicKey(context.Background())
	require.NoError(t, err)

	// Corrupt the signature (flip a byte).
	sig[10] ^= 0xFF

	resp := &ChallengeResponse{
		Nonce:     nonce,
		Signature: sig,
		PublicKey: pubkey,
		DID:       "did:lango:test",
	}

	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.Error(t, err)
}

func TestVerifyChallengeSignature_Roundtrip(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := ethcrypto.FromECDSA(privKey)

	s := &mockSigner{privKeyBytes: privBytes}
	h := newTestHandshaker(t, s)

	pubkey, err := s.PublicKey(context.Background())
	require.NoError(t, err)

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	timestamp := int64(1700000000)
	senderDID := "did:lango:abc123"

	// Sign the canonical payload (signer hashes with Keccak256 internally).
	canonical := challengeCanonicalPayload(nonce, timestamp, senderDID)
	sig, err := s.SignMessage(context.Background(), canonical)
	require.NoError(t, err)

	challenge := &Challenge{
		Nonce:     nonce,
		Timestamp: timestamp,
		SenderDID: senderDID,
		PublicKey: pubkey,
		Signature: sig,
	}

	// Verification should succeed (single hash on each side).
	err = h.verifyChallengeSignature(challenge)
	assert.NoError(t, err)
}

func TestVerifyChallengeSignature_WrongKey(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := ethcrypto.FromECDSA(privKey)
	s := &mockSigner{privKeyBytes: privBytes}
	h := newTestHandshaker(t, s)

	otherKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	otherPubkey := ethcrypto.CompressPubkey(&otherKey.PublicKey)

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	canonical := challengeCanonicalPayload(nonce, int64(1700000000), "did:lango:abc")
	sig, err := s.SignMessage(context.Background(), canonical)
	require.NoError(t, err)

	challenge := &Challenge{
		Nonce:     nonce,
		Timestamp: int64(1700000000),
		SenderDID: "did:lango:abc",
		PublicKey: otherPubkey,
		Signature: sig,
	}

	err = h.verifyChallengeSignature(challenge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public key mismatch")
}

func TestVerifyResponse_Ed25519(t *testing.T) {
	t.Parallel()

	// Ed25519 key pair for framework verification test.
	pub, priv, err := ed25519GenerateKey()
	require.NoError(t, err)

	s := &mockEd25519Signer{pub: pub, priv: priv}
	sessions, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	h := NewHandshaker(Config{
		Signer:   s,
		Sessions: sessions,
		Timeout:  30 * time.Second,
		Logger:   zap.NewNop().Sugar(),
		Verifiers: map[string]SignatureVerifyFunc{
			security.AlgorithmSecp256k1Keccak256: VerifySecp256k1Signature,
			security.AlgorithmEd25519:            security.VerifyEd25519,
		},
	})

	nonce := []byte("test-ed25519-nonce-32-bytes!!!!!!")
	sig := ed25519Sign(priv, nonce)

	resp := &ChallengeResponse{
		Nonce:              nonce,
		Signature:          sig,
		PublicKey:          pub,
		DID:                "did:lango:test-ed25519",
		SignatureAlgorithm: security.AlgorithmEd25519,
	}

	// NOTE: test-only did:lango:<ed25519-pubkey-hex> does NOT represent
	// production capability. Phase 3 DID v2 is required for Ed25519 DIDs.
	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.NoError(t, err)
}
