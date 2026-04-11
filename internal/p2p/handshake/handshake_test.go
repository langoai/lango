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
	canonical := challengeCanonicalPayload(nonce, timestamp, senderDID, "", nil)
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
	canonical := challengeCanonicalPayload(nonce, int64(1700000000), "did:lango:abc", "", nil)
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

// --- Phase 4: KEM-specific tests ---

func TestKEMHandshakeRoundtrip(t *testing.T) {
	t.Parallel()

	// Generate initiator and responder signers.
	initKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	respKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)

	initSigner := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(initKey)}
	respSigner := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(respKey)}

	initDID, _ := initSigner.DID(context.Background())
	respDID, _ := respSigner.DID(context.Background())

	// Initiator: generate KEM keypair.
	kemPub, kemDecap, err := security.GenerateEphemeralKEM()
	require.NoError(t, err)

	// Responder: encapsulate.
	ct, ssResp, err := security.KEMEncapsulate(kemPub)
	require.NoError(t, err)

	// Initiator: decapsulate.
	ssInit, err := kemDecap(ct)
	require.NoError(t, err)

	// Shared secrets must match.
	assert.Equal(t, ssResp, ssInit)

	// Derive session keys on both sides.
	keyInit, err := security.DeriveSessionKey(ssInit, initDID, respDID)
	require.NoError(t, err)

	keyResp, err := security.DeriveSessionKey(ssResp, initDID, respDID)
	require.NoError(t, err)

	// Session keys must be identical.
	assert.Equal(t, keyInit, keyResp)
	assert.Len(t, keyInit, 32)
}

func TestKEMGracefulDegradation(t *testing.T) {
	t.Parallel()

	// v1.2 initiator with KEM, v1.1 responder without KEM.
	initKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	initSigner := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(initKey)}

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")

	// Initiator generates KEM keypair.
	kemPub, _, err := security.GenerateEphemeralKEM()
	require.NoError(t, err)

	// Build challenge with KEM fields.
	initDID, _ := initSigner.DID(context.Background())
	challenge := Challenge{
		Nonce:        nonce,
		Timestamp:    time.Now().Unix(),
		SenderDID:    initDID,
		KEMPublicKey: kemPub,
		KEMAlgorithm: security.AlgorithmX25519MLKEM768,
	}

	// v1.1 responder signs nonce only (no KEM ciphertext).
	respKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	respSigner := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(respKey)}

	respPub, _ := respSigner.PublicKey(context.Background())
	// v1.1 response: sign nonce only (empty kemCiphertext → responseCanonicalPayload = nonce).
	signPayload := responseCanonicalPayload(challenge.Nonce, nil)
	sig, err := respSigner.SignMessage(context.Background(), signPayload)
	require.NoError(t, err)

	resp := &ChallengeResponse{
		Nonce:     nonce,
		Signature: sig,
		PublicKey: respPub,
		DID:       "did:lango:v1-responder",
		// No KEMCiphertext — v1.1 responder.
	}

	// Verify response succeeds (graceful degradation).
	sessions, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)
	h := NewHandshaker(Config{
		Signer:      initSigner,
		Sessions:    sessions,
		Timeout:     30 * time.Second,
		EnablePQKEM: true,
		Logger:      zap.NewNop().Sugar(),
	})

	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.NoError(t, err)

	// No KEM ciphertext means KEMUsed should be false.
	assert.Empty(t, resp.KEMCiphertext)
}

func TestSessionKeyZeroed(t *testing.T) {
	t.Parallel()

	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	sess, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)
	sess.EncryptionKey = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	sess.KEMUsed = true

	// Keep a reference to verify zeroing.
	keyRef := sess.EncryptionKey

	// Remove should zero the key.
	store.Remove("did:lango:peer1")

	// All bytes should be zero.
	for i, b := range keyRef {
		assert.Equalf(t, byte(0), b, "byte %d should be zero after Remove", i)
	}
}

func TestSessionKeyZeroedOnOverwrite(t *testing.T) {
	t.Parallel()

	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	// Create first session with encryption key.
	sess1, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)
	sess1.EncryptionKey = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	keyRef := sess1.EncryptionKey

	// Create second session for same peer (overwrite).
	_, err = store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	// First session's key should be zeroed.
	for i, b := range keyRef {
		assert.Equalf(t, byte(0), b, "byte %d should be zero after overwrite", i)
	}
}

func TestKEMTranscriptBinding(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	s := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(privKey)}

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	kemPub, _, err := security.GenerateEphemeralKEM()
	require.NoError(t, err)

	timestamp := time.Now().Unix()
	did, _ := s.DID(context.Background())

	// Sign canonical payload including KEM fields.
	canonical := challengeCanonicalPayload(nonce, timestamp, did, security.AlgorithmX25519MLKEM768, kemPub)
	sig, err := s.SignMessage(context.Background(), canonical)
	require.NoError(t, err)

	pubkey, _ := s.PublicKey(context.Background())

	sessions, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)
	h := NewHandshaker(Config{
		Signer:   s,
		Sessions: sessions,
		Timeout:  30 * time.Second,
		Logger:   zap.NewNop().Sugar(),
	})

	// Valid challenge with correct KEM fields.
	challenge := &Challenge{
		Nonce:              nonce,
		Timestamp:          timestamp,
		SenderDID:          did,
		PublicKey:          pubkey,
		Signature:          sig,
		SignatureAlgorithm: security.AlgorithmSecp256k1Keccak256,
		KEMPublicKey:       kemPub,
		KEMAlgorithm:       security.AlgorithmX25519MLKEM768,
	}
	err = h.verifyChallengeSignature(challenge)
	assert.NoError(t, err, "valid KEM challenge should pass")

	// Tampered KEM public key should fail signature verification.
	tamperedPub := make([]byte, len(kemPub))
	copy(tamperedPub, kemPub)
	tamperedPub[0] ^= 0xFF
	tamperedChallenge := *challenge
	tamperedChallenge.KEMPublicKey = tamperedPub
	err = h.verifyChallengeSignature(&tamperedChallenge)
	assert.Error(t, err, "tampered KEM public key should fail verification")
}

func TestResponseTranscriptBinding(t *testing.T) {
	t.Parallel()

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	s := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(privKey)}

	nonce := []byte("test-challenge-nonce-32bytes!!!!!")
	kemCt := []byte("fake-kem-ciphertext-for-binding-test")

	// Sign response payload: nonce || kemCiphertext.
	signPayload := responseCanonicalPayload(nonce, kemCt)
	sig, err := s.SignMessage(context.Background(), signPayload)
	require.NoError(t, err)

	pubkey, _ := s.PublicKey(context.Background())

	sessions, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)
	h := NewHandshaker(Config{
		Signer:   s,
		Sessions: sessions,
		Timeout:  30 * time.Second,
		Logger:   zap.NewNop().Sugar(),
	})

	// Valid response with matching ciphertext.
	resp := &ChallengeResponse{
		Nonce:         nonce,
		Signature:     sig,
		PublicKey:     pubkey,
		DID:           "did:lango:resp",
		KEMCiphertext: kemCt,
	}
	err = h.verifyResponse(context.Background(), resp, nonce)
	assert.NoError(t, err, "valid response with KEM ciphertext should pass")

	// Tampered ciphertext should fail.
	tamperedCt := make([]byte, len(kemCt))
	copy(tamperedCt, kemCt)
	tamperedCt[0] ^= 0xFF
	tamperedResp := *resp
	tamperedResp.KEMCiphertext = tamperedCt
	err = h.verifyResponse(context.Background(), &tamperedResp, nonce)
	assert.Error(t, err, "tampered KEM ciphertext should fail verification")
}

func TestPreferredProtocols(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		kemEnabled bool
		wantFirst  string
		wantLen    int
	}{
		{"KEM enabled", true, ProtocolIDv12, 3},
		{"KEM disabled", false, ProtocolIDv11, 2},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			protocols := PreferredProtocols(tt.kemEnabled)
			assert.Len(t, protocols, tt.wantLen)
			assert.Equal(t, tt.wantFirst, protocols[0])
		})
	}
}
