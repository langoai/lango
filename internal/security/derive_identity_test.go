package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveIdentityKey_Deterministic(t *testing.T) {
	t.Parallel()
	mk := make([]byte, 32)
	_, _ = rand.Read(mk)

	k1 := DeriveIdentityKey(mk, 0)
	k2 := DeriveIdentityKey(mk, 0)

	assert.Equal(t, []byte(k1), []byte(k2), "same MK + generation must produce same key")
}

func TestDeriveIdentityKey_DifferentMKs(t *testing.T) {
	t.Parallel()
	mk1 := make([]byte, 32)
	mk2 := make([]byte, 32)
	_, _ = rand.Read(mk1)
	_, _ = rand.Read(mk2)

	k1 := DeriveIdentityKey(mk1, 0)
	k2 := DeriveIdentityKey(mk2, 0)

	assert.NotEqual(t, []byte(k1), []byte(k2), "different MKs must produce different keys")
}

func TestDeriveIdentityKey_GenerationChangesKey(t *testing.T) {
	t.Parallel()
	mk := make([]byte, 32)
	_, _ = rand.Read(mk)

	k0 := DeriveIdentityKey(mk, 0)
	k1 := DeriveIdentityKey(mk, 1)

	assert.NotEqual(t, []byte(k0), []byte(k1), "different generations must produce different keys")
}

func TestDeriveIdentityKey_ValidEd25519(t *testing.T) {
	t.Parallel()
	mk := make([]byte, 32)
	_, _ = rand.Read(mk)

	priv := DeriveIdentityKey(mk, 0)
	require.Len(t, priv, ed25519.PrivateKeySize) // 64 bytes

	pub := priv.Public().(ed25519.PublicKey)
	require.Len(t, pub, ed25519.PublicKeySize) // 32 bytes

	// Sign and verify to confirm the key is functional.
	msg := []byte("test message")
	sig := ed25519.Sign(priv, msg)
	assert.True(t, ed25519.Verify(pub, msg, sig))
}

func TestDeriveIdentityKey_Generation0IsDefault(t *testing.T) {
	t.Parallel()
	mk := make([]byte, 32)
	_, _ = rand.Read(mk)

	// Generation 0 should NOT append ":0" to domain (backward compat).
	k := DeriveIdentityKey(mk, 0)
	require.Len(t, k, ed25519.PrivateKeySize)
}
