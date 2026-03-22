package provenance

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/p2p/identity"
)

func testSigner(t *testing.T) (string, BundleSignFunc) {
	t.Helper()
	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&key.PublicKey))
	require.NoError(t, err)
	return did.ID, func(_ context.Context, payload []byte) ([]byte, error) {
		return ethcrypto.Sign(ethcrypto.Keccak256(payload), key)
	}
}

func TestBundleService_ExportImportVerify(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc)

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

	did, signFn := testSigner(t)
	bundle, data, err := bundleSvc.Export(ctx, "sess-1", RedactionContent, did, signFn)
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
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc)

	did, signFn := testSigner(t)

	_, _, err := bundleSvc.Export(context.Background(), "sess-1", RedactionLevel("invalid"), did, signFn)
	require.ErrorIs(t, err, ErrInvalidRedaction)
}

func TestBundleService_Export_ValidRedactionLevels(t *testing.T) {
	cpStore := NewMemoryStore()
	treeStore := NewMemoryTreeStore()
	attrStore := NewMemoryAttributionStore()
	attrSvc := NewAttributionService(attrStore, cpStore, &stubTokenReader{})
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc)

	did, signFn := testSigner(t)
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
			bundle, _, err := bundleSvc.Export(ctx, "sess-1", tt.give, did, signFn)
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
	bundleSvc := NewBundleService(cpStore, treeStore, attrStore, attrSvc)

	bundle := &ProvenanceBundle{
		Version:        "1",
		RedactionLevel: RedactionLevel("bogus"),
		SignerDID:      "did:example:123",
	}
	err := bundleSvc.Verify(bundle)
	require.ErrorIs(t, err, ErrInvalidRedaction)
}
