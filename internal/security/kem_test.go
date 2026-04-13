package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKEMRoundtrip(t *testing.T) {
	// Generate ephemeral keypair.
	pubBytes, decap, err := GenerateEphemeralKEM()
	require.NoError(t, err)
	require.NotEmpty(t, pubBytes)
	require.NotNil(t, decap)

	// Encapsulate using the public key.
	ct, ssEncap, err := KEMEncapsulate(pubBytes)
	require.NoError(t, err)
	require.NotEmpty(t, ct)
	require.Len(t, ssEncap, HybridSharedSecretSize)

	// Decapsulate using the closure.
	ssDecap, err := decap(ct)
	require.NoError(t, err)
	require.Len(t, ssDecap, HybridSharedSecretSize)

	// Shared secrets must match.
	assert.Equal(t, ssEncap, ssDecap)
}

func TestKEMDifferentKeypairsProduceDifferentSecrets(t *testing.T) {
	pub1, _, err := GenerateEphemeralKEM()
	require.NoError(t, err)

	pub2, _, err := GenerateEphemeralKEM()
	require.NoError(t, err)

	// Different keypairs produce different public keys.
	assert.NotEqual(t, pub1, pub2)
}

func TestKEMEncapsulateInvalidPubkey(t *testing.T) {
	_, _, err := KEMEncapsulate([]byte("invalid"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal KEM public key")
}

func TestKEMDecapsulateInvalidCiphertext(t *testing.T) {
	_, decap, err := GenerateEphemeralKEM()
	require.NoError(t, err)

	_, err = decap([]byte("invalid-ciphertext"))
	require.Error(t, err)
}

func TestKEMPublicKeySize(t *testing.T) {
	scheme := HybridKEMScheme()
	pubBytes, _, err := GenerateEphemeralKEM()
	require.NoError(t, err)
	assert.Equal(t, scheme.PublicKeySize(), len(pubBytes))
}

func TestKEMCiphertextSize(t *testing.T) {
	scheme := HybridKEMScheme()
	pubBytes, _, err := GenerateEphemeralKEM()
	require.NoError(t, err)

	ct, _, err := KEMEncapsulate(pubBytes)
	require.NoError(t, err)
	assert.Equal(t, scheme.CiphertextSize(), len(ct))
}
