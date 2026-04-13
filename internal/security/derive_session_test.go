package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveSessionKeyDeterministic(t *testing.T) {
	ss := make([]byte, HybridSharedSecretSize)
	for i := range ss {
		ss[i] = byte(i)
	}

	key1, err := DeriveSessionKey(ss, "did:lango:alice", "did:lango:bob")
	require.NoError(t, err)
	require.Len(t, key1, KeySize)

	key2, err := DeriveSessionKey(ss, "did:lango:alice", "did:lango:bob")
	require.NoError(t, err)

	assert.Equal(t, key1, key2, "same inputs must produce same key")
}

func TestDeriveSessionKeyDifferentDIDs(t *testing.T) {
	ss := make([]byte, HybridSharedSecretSize)
	for i := range ss {
		ss[i] = byte(i)
	}

	key1, err := DeriveSessionKey(ss, "did:lango:alice", "did:lango:bob")
	require.NoError(t, err)

	key2, err := DeriveSessionKey(ss, "did:lango:alice", "did:lango:charlie")
	require.NoError(t, err)

	assert.NotEqual(t, key1, key2, "different DIDs must produce different keys")
}

func TestDeriveSessionKeyDIDOrderMatters(t *testing.T) {
	ss := make([]byte, HybridSharedSecretSize)
	for i := range ss {
		ss[i] = byte(i)
	}

	key1, err := DeriveSessionKey(ss, "did:lango:alice", "did:lango:bob")
	require.NoError(t, err)

	key2, err := DeriveSessionKey(ss, "did:lango:bob", "did:lango:alice")
	require.NoError(t, err)

	assert.NotEqual(t, key1, key2, "swapped DID order must produce different keys")
}

func TestDeriveSessionKeyWrongSharedSecretSize(t *testing.T) {
	tests := []struct {
		give string
		size int
	}{
		{"too short", 32},
		{"too long", 128},
		{"empty", 0},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			ss := make([]byte, tt.size)
			_, err := DeriveSessionKey(ss, "did:lango:a", "did:lango:b")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unexpected shared secret size")
		})
	}
}

func TestDeriveSessionKeyDifferentSharedSecrets(t *testing.T) {
	ss1 := make([]byte, HybridSharedSecretSize)
	ss2 := make([]byte, HybridSharedSecretSize)
	ss2[0] = 1

	key1, err := DeriveSessionKey(ss1, "did:lango:a", "did:lango:b")
	require.NoError(t, err)

	key2, err := DeriveSessionKey(ss2, "did:lango:a", "did:lango:b")
	require.NoError(t, err)

	assert.NotEqual(t, key1, key2, "different shared secrets must produce different keys")
}
