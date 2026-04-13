package identity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBundle() *IdentityBundle {
	return &IdentityBundle{
		Version: 1,
		SigningKey: PublicKeyEntry{
			Algorithm: "ed25519",
			PublicKey: make([]byte, 32),
		},
		SettlementKey: PublicKeyEntry{
			Algorithm: "secp256k1-keccak256",
			PublicKey: make([]byte, 33),
		},
		LegacyDID: "did:lango:abc123",
		CreatedAt: time.Now(),
	}
}

func TestBundleFile_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	bundle := testBundle()

	require.NoError(t, StoreBundleFile(dir, bundle))
	assert.True(t, HasBundleFile(dir))

	loaded, err := LoadBundleFile(dir)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, bundle.Version, loaded.Version)
	assert.Equal(t, bundle.SigningKey.Algorithm, loaded.SigningKey.Algorithm)
	assert.Equal(t, bundle.LegacyDID, loaded.LegacyDID)
}

func TestBundleFile_Missing(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, HasBundleFile(dir))

	loaded, err := LoadBundleFile(dir)
	assert.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestBundleFile_Nil(t *testing.T) {
	dir := t.TempDir()
	err := StoreBundleFile(dir, nil)
	assert.Error(t, err)
}

func TestKnownBundle_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	bundle := testBundle()
	didV2 := "did:lango:v2:abcdef1234567890abcdef1234567890abcdef12"

	require.NoError(t, StoreKnownBundle(dir, didV2, bundle))

	loaded, err := LoadKnownBundle(dir, didV2)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, bundle.Version, loaded.Version)
}

func TestKnownBundle_Missing(t *testing.T) {
	dir := t.TempDir()
	loaded, err := LoadKnownBundle(dir, "did:lango:v2:missing")
	assert.NoError(t, err)
	assert.Nil(t, loaded)
}
