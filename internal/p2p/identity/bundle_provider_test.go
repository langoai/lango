package identity

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleProvider_Creation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	walletKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	walletPub := ethcrypto.CompressPubkey(&walletKey.PublicKey)

	legacyProv := NewProvider(&mockKeyProvider{pubkey: walletPub}, testLogger())

	bp, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:     priv,
		SettlementPub: walletPub,
		LangoDir:      dir,
		Legacy:        legacyProv,
		Logger:        testLogger(),
	})
	require.NoError(t, err)

	// Bundle should be created.
	bundle := bp.Bundle()
	require.NotNil(t, bundle)
	assert.Equal(t, "ed25519", bundle.SigningKey.Algorithm)
	assert.Equal(t, "secp256k1-keccak256", bundle.SettlementKey.Algorithm)
	assert.NotEmpty(t, bundle.LegacyDID)

	// DID should be v2.
	did, err := bp.DID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, did.Version)
	assert.True(t, strings.HasPrefix(did.ID, "did:lango:v2:"))
	assert.NotEmpty(t, did.PeerID)

	// Ed25519 proof should be present.
	assert.NotEmpty(t, bundle.Proofs.Ed25519)

	// Bundle file should be persisted.
	assert.True(t, HasBundleFile(dir))
}

func TestBundleProvider_LoadExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	walletKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	walletPub := ethcrypto.CompressPubkey(&walletKey.PublicKey)

	legacyProv := NewProvider(&mockKeyProvider{pubkey: walletPub}, testLogger())

	// Create first.
	bp1, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:     priv,
		SettlementPub: walletPub,
		LangoDir:      dir,
		Legacy:        legacyProv,
		Logger:        testLogger(),
	})
	require.NoError(t, err)
	did1, err := bp1.DID(context.Background())
	require.NoError(t, err)

	// Load again — should reuse existing.
	bp2, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:     priv,
		SettlementPub: walletPub,
		LangoDir:      dir,
		Legacy:        legacyProv,
		Logger:        testLogger(),
	})
	require.NoError(t, err)
	did2, err := bp2.DID(context.Background())
	require.NoError(t, err)

	assert.Equal(t, did1.ID, did2.ID)
}

func TestBundleProvider_SignMessage(t *testing.T) {
	t.Parallel()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	pub := priv.Public().(ed25519.PublicKey)

	bp, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:     priv,
		SettlementPub: make([]byte, 33),
		Logger:        testLogger(),
	})
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := bp.SignMessage(context.Background(), msg)
	require.NoError(t, err)

	assert.True(t, ed25519.Verify(pub, msg, sig))
}

func TestBundleProvider_Algorithm(t *testing.T) {
	t.Parallel()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	bp, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:     priv,
		SettlementPub: make([]byte, 33),
		Logger:        testLogger(),
	})
	require.NoError(t, err)

	assert.Equal(t, "ed25519", bp.Algorithm())
}

func TestBundleProvider_LegacyDID(t *testing.T) {
	t.Parallel()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	walletKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	walletPub := ethcrypto.CompressPubkey(&walletKey.PublicKey)

	legacyProv := NewProvider(&mockKeyProvider{pubkey: walletPub}, testLogger())

	bp, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:     priv,
		SettlementPub: walletPub,
		Legacy:        legacyProv,
		Logger:        testLogger(),
	})
	require.NoError(t, err)

	legacyDID, err := bp.LegacyDID(context.Background())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(legacyDID.ID, "did:lango:"))
	assert.Equal(t, 1, legacyDID.Version)
}

func TestBundleProvider_DIDString(t *testing.T) {
	t.Parallel()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	bp, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:     priv,
		SettlementPub: make([]byte, 33),
		Logger:        testLogger(),
	})
	require.NoError(t, err)

	didStr, err := bp.DIDString(context.Background())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(didStr, "did:lango:v2:"))
}

func TestBundleProvider_NilArgs(t *testing.T) {
	t.Parallel()

	_, err := NewBundleProvider(BundleProviderConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing key and settlement public key are required")
}
