package security

import (
	"testing"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMLDSA65SignVerifyRoundtrip(t *testing.T) {
	pk, sk, err := mldsa65.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("hello post-quantum world")
	sig, err := SignMLDSA65(sk, message)
	require.NoError(t, err)
	require.Len(t, sig, mldsa65.SignatureSize)

	pubBytes, _ := pk.MarshalBinary()
	err = VerifyMLDSA65(pubBytes, message, sig)
	assert.NoError(t, err)
}

func TestMLDSA65VerifyWrongMessage(t *testing.T) {
	pk, sk, err := mldsa65.GenerateKey(nil)
	require.NoError(t, err)

	sig, err := SignMLDSA65(sk, []byte("original"))
	require.NoError(t, err)

	pubBytes, _ := pk.MarshalBinary()
	err = VerifyMLDSA65(pubBytes, []byte("tampered"), sig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")
}

func TestMLDSA65VerifyInvalidPubkey(t *testing.T) {
	err := VerifyMLDSA65([]byte("short"), []byte("msg"), make([]byte, mldsa65.SignatureSize))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key length")
}

func TestMLDSA65VerifyInvalidSignatureLength(t *testing.T) {
	pubBytes := make([]byte, mldsa65.PublicKeySize)
	err := VerifyMLDSA65(pubBytes, []byte("msg"), []byte("short"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature length")
}

func TestMLDSA65SchemeMetadata(t *testing.T) {
	assert.Equal(t, AlgorithmMLDSA65, MLDSA65Scheme.ID)
	assert.Equal(t, mldsa65.SignatureSize, MLDSA65Scheme.SignatureSize)
	assert.Equal(t, mldsa65.PublicKeySize, MLDSA65Scheme.PublicKeySize)
	assert.NotNil(t, MLDSA65Scheme.Verify)
}
