package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecp256k1Keccak256Scheme_Metadata(t *testing.T) {
	t.Parallel()
	assert.Equal(t, AlgorithmSecp256k1Keccak256, Secp256k1Keccak256Scheme.ID)
	assert.Equal(t, 65, Secp256k1Keccak256Scheme.SignatureSize)
	assert.Equal(t, 33, Secp256k1Keccak256Scheme.PublicKeySize)
}

func TestEd25519Scheme_Metadata(t *testing.T) {
	t.Parallel()
	assert.Equal(t, AlgorithmEd25519, Ed25519Scheme.ID)
	assert.Equal(t, 64, Ed25519Scheme.SignatureSize)
	assert.Equal(t, 32, Ed25519Scheme.PublicKeySize)
}

func TestVerifySecp256k1Keccak256_ValidSignature(t *testing.T) {
	t.Parallel()

	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)

	pubkey := ethcrypto.CompressPubkey(&key.PublicKey)
	message := []byte("test message for secp256k1")
	hash := ethcrypto.Keccak256(message)
	sig, err := ethcrypto.Sign(hash, key)
	require.NoError(t, err)

	err = VerifySecp256k1Keccak256(pubkey, message, sig)
	assert.NoError(t, err)
}

func TestVerifySecp256k1Keccak256_InvalidSignature(t *testing.T) {
	t.Parallel()

	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)

	otherKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	otherPubkey := ethcrypto.CompressPubkey(&otherKey.PublicKey)

	message := []byte("test message")
	hash := ethcrypto.Keccak256(message)
	sig, err := ethcrypto.Sign(hash, key)
	require.NoError(t, err)

	err = VerifySecp256k1Keccak256(otherPubkey, message, sig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public key mismatch")
}

func TestVerifySecp256k1Keccak256_WrongSigLength(t *testing.T) {
	t.Parallel()
	err := VerifySecp256k1Keccak256(make([]byte, 33), []byte("msg"), []byte("short"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature length")
}

func TestVerifySecp256k1Keccak256_WrongKeyLength(t *testing.T) {
	t.Parallel()
	err := VerifySecp256k1Keccak256(make([]byte, 32), []byte("msg"), make([]byte, 65))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key length")
}

func TestVerifyEd25519_ValidSignature(t *testing.T) {
	t.Parallel()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	message := []byte("test message for ed25519")
	sig := ed25519.Sign(priv, message)

	err = VerifyEd25519(pub, message, sig)
	assert.NoError(t, err)
}

func TestVerifyEd25519_InvalidSignature(t *testing.T) {
	t.Parallel()

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	_, otherPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	message := []byte("test message")
	sig := ed25519.Sign(otherPriv, message)

	err = VerifyEd25519(pub, message, sig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")
}

func TestVerifyEd25519_WrongKeyLength(t *testing.T) {
	t.Parallel()
	err := VerifyEd25519(make([]byte, 33), []byte("msg"), make([]byte, 64))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key length")
}

func TestVerifyEd25519_WrongSigLength(t *testing.T) {
	t.Parallel()
	err := VerifyEd25519(make([]byte, 32), []byte("msg"), make([]byte, 65))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature length")
}
