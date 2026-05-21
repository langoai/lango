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

	"github.com/langoai/lango/internal/security"
)

var bundleProviderTestSeed = []byte{
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
	0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
	0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
}

func newDeterministicBundleProvider(t *testing.T, cfg BundleProviderConfig) *BundleProvider {
	t.Helper()

	if cfg.SigningKey == nil {
		cfg.SigningKey = ed25519.NewKeyFromSeed(bundleProviderTestSeed)
	}
	if len(cfg.SettlementPub) == 0 {
		cfg.SettlementPub = make([]byte, 33)
	}
	if cfg.Logger == nil {
		cfg.Logger = testLogger()
	}

	bp, err := NewBundleProvider(cfg)
	require.NoError(t, err)
	return bp
}

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
		SigningKey:    priv,
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
		SigningKey:    priv,
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
		SigningKey:    priv,
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
		SigningKey:    priv,
		SettlementPub: make([]byte, 33),
		Logger:        testLogger(),
	})
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := bp.SignMessage(context.Background(), msg)
	require.NoError(t, err)

	assert.True(t, ed25519.Verify(pub, msg, sig))
}

func TestBundleProvider_PublicKeyAccessorReturnsEd25519PublicKey(t *testing.T) {
	t.Parallel()

	priv := ed25519.NewKeyFromSeed(bundleProviderTestSeed)
	wantPub := priv.Public().(ed25519.PublicKey)

	bp := newDeterministicBundleProvider(t, BundleProviderConfig{
		SigningKey: priv,
	})

	gotPub, err := bp.PublicKey(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte(wantPub), gotPub)
}

func TestBundleProvider_Algorithm(t *testing.T) {
	t.Parallel()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	bp, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:    priv,
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
		SigningKey:    priv,
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

func TestBundleProvider_LegacyDIDWithoutLegacyProvider(t *testing.T) {
	t.Parallel()

	bp := newDeterministicBundleProvider(t, BundleProviderConfig{})

	legacyDID, err := bp.LegacyDID(context.Background())
	require.Error(t, err)
	assert.Nil(t, legacyDID)
	assert.Contains(t, err.Error(), "no legacy identity provider")
}

func TestBundleProvider_DIDString(t *testing.T) {
	t.Parallel()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	bp, err := NewBundleProvider(BundleProviderConfig{
		SigningKey:    priv,
		SettlementPub: make([]byte, 33),
		Logger:        testLogger(),
	})
	require.NoError(t, err)

	didStr, err := bp.DIDString(context.Background())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(didStr, "did:lango:v2:"))
}

func TestBundleProvider_VerifyDIDRejectsNilAndV2(t *testing.T) {
	t.Parallel()

	bp := newDeterministicBundleProvider(t, BundleProviderConfig{})

	err := bp.VerifyDID(nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil DID")

	err = bp.VerifyDID(&DID{Version: 2}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "v2 DID verification requires BundleResolver")
}

func TestBundleProvider_VerifyDIDV1WithoutLegacyProvider(t *testing.T) {
	t.Parallel()

	bp := newDeterministicBundleProvider(t, BundleProviderConfig{})

	err := bp.VerifyDID(&DID{Version: 1}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no legacy provider for v1 DID verification")
}

func TestBundleProvider_PQMethodsWithoutPQKey(t *testing.T) {
	t.Parallel()

	bp := newDeterministicBundleProvider(t, BundleProviderConfig{})

	sig, err := bp.SignPQ(context.Background(), []byte("message"))
	require.Error(t, err)
	assert.Nil(t, sig)
	assert.Contains(t, err.Error(), "PQ signing key not available")
	assert.Nil(t, bp.PQPublicKey())
	assert.False(t, bp.HasPQKey())
	assert.Equal(t, security.AlgorithmMLDSA65, bp.PQAlgorithm())
}

func TestBundleProvider_PQMethodsWithSeed(t *testing.T) {
	t.Parallel()

	bp := newDeterministicBundleProvider(t, BundleProviderConfig{
		PQSigningKeySeed: bundleProviderTestSeed,
	})

	bundle := bp.Bundle()
	require.NotNil(t, bundle.PQSigningKey)
	assert.Equal(t, security.AlgorithmMLDSA65, bundle.PQSigningKey.Algorithm)
	assert.NotEmpty(t, bundle.PQSigningKey.PublicKey)
	assert.NotEmpty(t, bundle.Proofs.MLDSA65)
	assert.True(t, bp.HasPQKey())

	pqPub := bp.PQPublicKey()
	assert.NotEmpty(t, pqPub)
	assert.Equal(t, bundle.PQSigningKey.PublicKey, pqPub)

	msg := []byte("deterministic pq message")
	sig, err := bp.SignPQ(context.Background(), msg)
	require.NoError(t, err)
	assert.NotEmpty(t, sig)
	assert.NoError(t, security.VerifyMLDSA65(pqPub, msg, sig))

	canonical, err := CanonicalBundleBytes(bundle)
	require.NoError(t, err)
	assert.NoError(t, security.VerifyMLDSA65(bundle.PQSigningKey.PublicKey, canonical, bundle.Proofs.MLDSA65))
}

func TestBundleProvider_NilArgs(t *testing.T) {
	t.Parallel()

	_, err := NewBundleProvider(BundleProviderConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing key and settlement public key are required")
}
