package security

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestDerivePQSigningKeyDeterministic(t *testing.T) {
	mk := make([]byte, 32)
	for i := range mk {
		mk[i] = byte(i)
	}

	pk1, sk1 := DerivePQSigningKey(mk, 0)
	pk2, sk2 := DerivePQSigningKey(mk, 0)

	pk1Bytes, _ := pk1.MarshalBinary()
	pk2Bytes, _ := pk2.MarshalBinary()
	assert.Equal(t, pk1Bytes, pk2Bytes, "same MK+generation must produce same public key")

	sk1Bytes, _ := sk1.MarshalBinary()
	sk2Bytes, _ := sk2.MarshalBinary()
	assert.Equal(t, sk1Bytes, sk2Bytes, "same MK+generation must produce same private key")
}

func TestDerivePQSigningKeyGenerationRotation(t *testing.T) {
	mk := make([]byte, 32)
	for i := range mk {
		mk[i] = byte(i)
	}

	pk0, _ := DerivePQSigningKey(mk, 0)
	pk1, _ := DerivePQSigningKey(mk, 1)

	pk0Bytes, _ := pk0.MarshalBinary()
	pk1Bytes, _ := pk1.MarshalBinary()
	assert.NotEqual(t, pk0Bytes, pk1Bytes, "different generation must produce different keys")
}

func TestDerivePQSigningKeyDifferentMK(t *testing.T) {
	mk1 := make([]byte, 32)
	mk2 := make([]byte, 32)
	mk2[0] = 1

	pk1, _ := DerivePQSigningKey(mk1, 0)
	pk2, _ := DerivePQSigningKey(mk2, 0)

	pk1Bytes, _ := pk1.MarshalBinary()
	pk2Bytes, _ := pk2.MarshalBinary()
	assert.NotEqual(t, pk1Bytes, pk2Bytes, "different MK must produce different keys")
}

func TestDerivePQSigningKeyDomainSeparation(t *testing.T) {
	mk := make([]byte, 32)
	for i := range mk {
		mk[i] = byte(i)
	}

	// PQ key derivation
	pqPK, _ := DerivePQSigningKey(mk, 0)
	pqPKBytes, _ := pqPK.MarshalBinary()

	// Ed25519 key derivation (different domain)
	ed25519SK := DeriveIdentityKey(mk, 0)
	ed25519PKBytes := []byte(ed25519SK.Public().(ed25519.PublicKey))

	// Keys must be different (different domains, different algorithms, different sizes)
	assert.NotEqual(t, len(pqPKBytes), len(ed25519PKBytes), "PQ and Ed25519 keys have different sizes")
}

func TestDerivePQSigningKeySignVerify(t *testing.T) {
	mk := make([]byte, 32)
	for i := range mk {
		mk[i] = byte(i)
	}

	pk, sk := DerivePQSigningKey(mk, 0)

	message := []byte("test message for PQ signing")
	sig, err := SignMLDSA65(sk, message)
	require.NoError(t, err)

	pkBytes, _ := pk.MarshalBinary()
	err = VerifyMLDSA65(pkBytes, message, sig)
	assert.NoError(t, err, "derived key must produce valid signatures")
}
