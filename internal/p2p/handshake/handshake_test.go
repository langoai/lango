package handshake

import (
	"context"
	"math/big"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockWallet implements wallet.WalletProvider for testing.
type mockWallet struct {
	privKeyBytes []byte
}

func (m *mockWallet) SignMessage(_ context.Context, message []byte) ([]byte, error) {
	key, err := ethcrypto.ToECDSA(m.privKeyBytes)
	if err != nil {
		return nil, err
	}
	hash := ethcrypto.Keccak256(message)
	return ethcrypto.Sign(hash, key)
}

func (m *mockWallet) PublicKey(_ context.Context) ([]byte, error) {
	key, err := ethcrypto.ToECDSA(m.privKeyBytes)
	if err != nil {
		return nil, err
	}
	return ethcrypto.CompressPubkey(&key.PublicKey), nil
}

func (m *mockWallet) Address(_ context.Context) (string, error)                   { return "", nil }
func (m *mockWallet) Balance(_ context.Context) (*big.Int, error)                 { return nil, nil }
func (m *mockWallet) SignTransaction(_ context.Context, _ []byte) ([]byte, error) { return nil, nil }

func newTestHandshaker(t *testing.T, w *mockWallet) *Handshaker {
	t.Helper()
	sessions, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	return NewHandshaker(Config{
		Wallet:   w,
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

	w := &mockWallet{privKeyBytes: privBytes}
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

	w := &mockWallet{privKeyBytes: privBytes}
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

	w := &mockWallet{privKeyBytes: privBytes}
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

	w := &mockWallet{privKeyBytes: privBytes}
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

	w := &mockWallet{privKeyBytes: privBytes}
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

	w := &mockWallet{privKeyBytes: privBytes}
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
