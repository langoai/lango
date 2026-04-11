package provenance

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/types"
)

// testBundleSigner implements BundleSigner for testing.
type testBundleSigner struct {
	key *ecdsa.PrivateKey
}

func (s *testBundleSigner) Sign(_ context.Context, payload []byte) ([]byte, error) {
	return ethcrypto.Sign(ethcrypto.Keccak256(payload), s.key)
}

func (s *testBundleSigner) Algorithm() string {
	return AlgorithmSecp256k1Keccak256
}

func testSignerAndDID(t *testing.T) (string, *testBundleSigner) {
	t.Helper()
	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&key.PublicKey))
	require.NoError(t, err)
	return did.ID, &testBundleSigner{key: key}
}

// testVerifiers returns the default verifier map for tests.
func testVerifiers() map[string]SignatureVerifyFunc {
	return map[string]SignatureVerifyFunc{
		AlgorithmSecp256k1Keccak256: identity.VerifyMessageSignature,
	}
}

func TestBundleService_ExportImportVerify(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc, testVerifiers())

	ctx := context.Background()
	nowTime := time.Now()
	require.NoError(t, cpStore.SaveCheckpoint(ctx, Checkpoint{
		ID:         "cp-1",
		SessionKey: "sess-1",
		Label:      "checkpoint",
		Trigger:    TriggerManual,
		CreatedAt:  nowTime,
		Metadata:   map[string]string{"a": "b"},
		GitRef:     "abc123",
	}))
	require.NoError(t, treeStore.SaveNode(ctx, SessionNode{
		SessionKey: "sess-1",
		AgentName:  "root",
		Depth:      0,
		Status:     SessionStatusActive,
		CreatedAt:  nowTime,
		Goal:       "ship provenance",
	}))
	require.NoError(t, attrSvc.RecordWorkspaceOperation(ctx, "sess-1", "", "ws-1", AuthorAgent, "operator", "deadbeef", "", AttributionSourceWorkspaceMerge, []GitFileStat{
		{FilePath: "main.go", LinesAdded: 3, LinesRemoved: 1},
	}))

	did, signer := testSignerAndDID(t)
	bundle, data, err := bundleSvc.Export(ctx, "sess-1", RedactionContent, did, signer)
	require.NoError(t, err)
	assert.Equal(t, did, bundle.SignerDID)
	require.NoError(t, bundleSvc.Verify(bundle))
	assert.Empty(t, bundle.Checkpoints[0].GitRef)
	assert.Empty(t, bundle.Attributions[0].FilePath)

	imported, err := bundleSvc.Import(ctx, data)
	require.NoError(t, err)
	assert.Equal(t, did, imported.SignerDID)

	var tampered ProvenanceBundle
	require.NoError(t, json.Unmarshal(data, &tampered))
	tampered.RedactionLevel = RedactionFull
	err = bundleSvc.Verify(&tampered)
	require.Error(t, err)
}

func TestBundleService_Export_InvalidRedaction(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc, testVerifiers())

	did, signer := testSignerAndDID(t)

	_, _, err := bundleSvc.Export(context.Background(), "sess-1", RedactionLevel("invalid"), did, signer)
	require.ErrorIs(t, err, ErrInvalidRedaction)
}

func TestBundleService_Export_ValidRedactionLevels(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc, testVerifiers())

	did, signer := testSignerAndDID(t)
	ctx := context.Background()

	tests := []struct {
		give RedactionLevel
	}{
		{give: RedactionNone},
		{give: RedactionContent},
		{give: RedactionFull},
	}
	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			bundle, _, err := bundleSvc.Export(ctx, "sess-1", tt.give, did, signer)
			require.NoError(t, err)
			assert.Equal(t, tt.give, bundle.RedactionLevel)
		})
	}
}

func TestBundleService_Verify_InvalidRedaction(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc, testVerifiers())

	bundle := &ProvenanceBundle{
		Version:        "1",
		RedactionLevel: RedactionLevel("bogus"),
		SignerDID:      "did:example:123",
	}
	err := bundleSvc.Verify(bundle)
	require.ErrorIs(t, err, ErrInvalidRedaction)
}

// ed25519BundleSigner implements BundleSigner for Ed25519 (framework testing only).
type ed25519BundleSigner struct {
	priv ed25519.PrivateKey
}

func (s *ed25519BundleSigner) Sign(_ context.Context, payload []byte) ([]byte, error) {
	return ed25519.Sign(s.priv, payload), nil
}

func (s *ed25519BundleSigner) Algorithm() string { return security.AlgorithmEd25519 }

func TestBundleService_ExportVerify_Ed25519(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})

	// Ed25519 verifier: same wiring closure pattern as production.
	ed25519Verifier := func(didStr string, payload, signature []byte) error {
		pubkey, err := identity.ParseDIDPublicKey(didStr)
		if err != nil {
			return err
		}
		return security.VerifyEd25519(pubkey, payload, signature)
	}
	verifiers := map[string]SignatureVerifyFunc{
		AlgorithmSecp256k1Keccak256:    identity.VerifyMessageSignature,
		security.AlgorithmEd25519: ed25519Verifier,
	}
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc, verifiers)

	ctx := context.Background()
	require.NoError(t, cpStore.SaveCheckpoint(ctx, Checkpoint{
		ID:         "cp-ed25519",
		SessionKey: "sess-ed25519",
		Label:      "ed25519-test",
		Trigger:    TriggerManual,
		CreatedAt:  time.Now(),
	}))

	// Generate Ed25519 key pair and construct a test DID.
	// NOTE: test-only did:lango:<ed25519-pubkey-hex> does NOT represent
	// production capability. Phase 3 DID v2 is required for Ed25519 DIDs.
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	testDID := types.DIDPrefix + hex.EncodeToString(pub)

	signer := &ed25519BundleSigner{priv: priv}
	bundle, _, err := bundleSvc.Export(ctx, "sess-ed25519", RedactionNone, testDID, signer)
	require.NoError(t, err)
	assert.Equal(t, security.AlgorithmEd25519, bundle.SignatureAlgorithm)
	assert.Equal(t, testDID, bundle.SignerDID)

	// Verify using the Ed25519 verifier.
	err = bundleSvc.Verify(bundle)
	assert.NoError(t, err)
}

// pqDualSigner implements both BundleSigner (Ed25519) and PQBundleSigner (ML-DSA-65).
type pqDualSigner struct {
	ed25519Priv ed25519.PrivateKey
	pqPriv      *mldsa65.PrivateKey
	pqPub       *mldsa65.PublicKey
}

func (s *pqDualSigner) Sign(_ context.Context, payload []byte) ([]byte, error) {
	return ed25519.Sign(s.ed25519Priv, payload), nil
}

func (s *pqDualSigner) Algorithm() string { return security.AlgorithmEd25519 }

func (s *pqDualSigner) SignPQ(_ context.Context, payload []byte) ([]byte, error) {
	return security.SignMLDSA65(s.pqPriv, payload)
}

func (s *pqDualSigner) PQAlgorithm() string { return security.AlgorithmMLDSA65 }

func (s *pqDualSigner) PQPublicKey() []byte {
	b, _ := s.pqPub.MarshalBinary()
	return b
}

func TestBundleService_DualSignature(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})

	ed25519Verifier := func(didStr string, payload, signature []byte) error {
		pubkey, err := identity.ParseDIDPublicKey(didStr)
		if err != nil {
			return err
		}
		return security.VerifyEd25519(pubkey, payload, signature)
	}
	verifiers := map[string]SignatureVerifyFunc{
		security.AlgorithmEd25519: ed25519Verifier,
	}
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc, verifiers)

	ctx := context.Background()
	require.NoError(t, cpStore.SaveCheckpoint(ctx, Checkpoint{
		ID:         "cp-dual",
		SessionKey: "sess-dual",
		Label:      "dual-sig-test",
		Trigger:    TriggerManual,
		CreatedAt:  time.Now(),
	}))

	// Generate Ed25519 + ML-DSA-65 key pairs.
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	pqPub, pqPriv, err := mldsa65.GenerateKey(nil)
	require.NoError(t, err)

	testDID := types.DIDPrefix + hex.EncodeToString(edPub)
	signer := &pqDualSigner{ed25519Priv: edPriv, pqPriv: pqPriv, pqPub: pqPub}

	// Export with dual signer.
	bundle, _, err := bundleSvc.Export(ctx, "sess-dual", RedactionNone, testDID, signer)
	require.NoError(t, err)

	// Verify classical fields.
	assert.Equal(t, security.AlgorithmEd25519, bundle.SignatureAlgorithm)
	assert.NotEmpty(t, bundle.Signature)

	// Verify PQ fields.
	assert.Equal(t, security.AlgorithmMLDSA65, bundle.PQSignatureAlgorithm)
	assert.NotEmpty(t, bundle.PQSignature)
	assert.NotEmpty(t, bundle.PQSignerPublicKey)

	// Dual verification.
	err = bundleSvc.Verify(bundle)
	assert.NoError(t, err, "dual-signed bundle should verify")
}

func TestBundleService_ClassicalOnlyBackwardCompat(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})

	ed25519Verifier := func(didStr string, payload, signature []byte) error {
		pubkey, err := identity.ParseDIDPublicKey(didStr)
		if err != nil {
			return err
		}
		return security.VerifyEd25519(pubkey, payload, signature)
	}
	verifiers := map[string]SignatureVerifyFunc{
		security.AlgorithmEd25519: ed25519Verifier,
	}
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc, verifiers)

	ctx := context.Background()
	require.NoError(t, cpStore.SaveCheckpoint(ctx, Checkpoint{
		ID:         "cp-classic",
		SessionKey: "sess-classic",
		Label:      "classic-only",
		Trigger:    TriggerManual,
		CreatedAt:  time.Now(),
	}))

	// Ed25519-only signer (no PQBundleSigner).
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	testDID := types.DIDPrefix + hex.EncodeToString(edPub)
	signer := &ed25519BundleSigner{priv: edPriv}

	bundle, _, err := bundleSvc.Export(ctx, "sess-classic", RedactionNone, testDID, signer)
	require.NoError(t, err)

	// PQ fields should be empty.
	assert.Empty(t, bundle.PQSignature)
	assert.Empty(t, bundle.PQSignerPublicKey)
	assert.Empty(t, bundle.PQSignatureAlgorithm)

	// Classical-only verification should pass.
	err = bundleSvc.Verify(bundle)
	assert.NoError(t, err, "classical-only bundle should verify")
}
