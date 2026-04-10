package provenance

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/p2p/identity"
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
