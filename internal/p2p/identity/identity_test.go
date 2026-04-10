package identity

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/types"
)

func testLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

// generateTestPubkey creates a compressed secp256k1 public key for testing.
func generateTestPubkey(t *testing.T) []byte {
	t.Helper()
	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	return ethcrypto.CompressPubkey(&key.PublicKey)
}

func TestDIDPrefix_Constant(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "did:lango:", types.DIDPrefix)
}

func TestDIDFromPublicKey_Valid(t *testing.T) {
	t.Parallel()

	pubkey := generateTestPubkey(t)

	did, err := DIDFromPublicKey(pubkey)
	require.NoError(t, err)
	require.NotNil(t, did)

	assert.True(t, strings.HasPrefix(did.ID, types.DIDPrefix))
	assert.Equal(t, pubkey, did.PublicKey)
	assert.NotEmpty(t, did.PeerID)

	// Verify the hex encoding in the DID string.
	hexPart := strings.TrimPrefix(did.ID, types.DIDPrefix)
	decoded, err := hex.DecodeString(hexPart)
	require.NoError(t, err)
	assert.Equal(t, pubkey, decoded)
}

func TestDIDFromPublicKey_EmptyKey(t *testing.T) {
	t.Parallel()

	did, err := DIDFromPublicKey(nil)
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "empty public key")

	did, err = DIDFromPublicKey([]byte{})
	assert.Error(t, err)
	assert.Nil(t, did)
}

func TestParseDID_Valid_Roundtrip(t *testing.T) {
	t.Parallel()

	pubkey := generateTestPubkey(t)

	original, err := DIDFromPublicKey(pubkey)
	require.NoError(t, err)

	parsed, err := ParseDID(original.ID)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	assert.Equal(t, original.ID, parsed.ID)
	assert.Equal(t, original.PublicKey, parsed.PublicKey)
	assert.Equal(t, original.PeerID, parsed.PeerID)
}

func TestParseDID_InvalidPrefix(t *testing.T) {
	t.Parallel()

	did, err := ParseDID("did:other:abc123")
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "invalid DID scheme")
}

func TestParseDID_EmptyKey(t *testing.T) {
	t.Parallel()

	did, err := ParseDID("did:lango:")
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "empty public key")
}

func TestParseDID_InvalidHex(t *testing.T) {
	t.Parallel()

	did, err := ParseDID("did:lango:ZZZZ_not_hex")
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "decode hex")
}

func TestParseDIDPublicKey_Valid(t *testing.T) {
	t.Parallel()

	pubkey := generateTestPubkey(t)
	did, err := DIDFromPublicKey(pubkey)
	require.NoError(t, err)

	extracted, err := ParseDIDPublicKey(did.ID)
	require.NoError(t, err)
	assert.Equal(t, pubkey, extracted)
}

func TestParseDIDPublicKey_InvalidPrefix(t *testing.T) {
	t.Parallel()
	_, err := ParseDIDPublicKey("did:other:abc123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid DID scheme")
}

func TestParseDIDPublicKey_EmptyKey(t *testing.T) {
	t.Parallel()
	_, err := ParseDIDPublicKey("did:lango:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty public key")
}

func TestParseDIDPublicKey_InvalidHex(t *testing.T) {
	t.Parallel()
	_, err := ParseDIDPublicKey("did:lango:ZZZZ_not_hex")
	assert.Error(t, err)
}

func TestVerifyDID_Matching(t *testing.T) {
	t.Parallel()

	pubkey := generateTestPubkey(t)
	did, err := DIDFromPublicKey(pubkey)
	require.NoError(t, err)

	provider := NewProvider(&mockKeyProvider{pubkey: pubkey}, testLogger())
	err = provider.VerifyDID(did, did.PeerID)
	assert.NoError(t, err)
}

func TestVerifyDID_Mismatched(t *testing.T) {
	t.Parallel()

	pubkey := generateTestPubkey(t)
	did, err := DIDFromPublicKey(pubkey)
	require.NoError(t, err)

	// Generate a different peer ID.
	otherPubkey := generateTestPubkey(t)
	otherDID, err := DIDFromPublicKey(otherPubkey)
	require.NoError(t, err)

	provider := NewProvider(&mockKeyProvider{pubkey: pubkey}, testLogger())
	err = provider.VerifyDID(did, otherDID.PeerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "peer ID mismatch")
}

func TestVerifyDID_NilDID(t *testing.T) {
	t.Parallel()

	provider := NewProvider(&mockKeyProvider{}, testLogger())
	err := provider.VerifyDID(nil, peer.ID("somepeerid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil DID")
}

func TestWalletDIDProvider_DID_Caching(t *testing.T) {
	t.Parallel()

	pubkey := generateTestPubkey(t)
	mock := &mockKeyProvider{pubkey: pubkey}
	provider := NewProvider(mock, testLogger())

	did1, err := provider.DID(context.Background())
	require.NoError(t, err)

	did2, err := provider.DID(context.Background())
	require.NoError(t, err)

	assert.Same(t, did1, did2, "second call should return cached DID")
	assert.Equal(t, 1, mock.calls, "PublicKey should only be called once due to caching")
}

func TestWalletDIDProvider_DID_WalletError(t *testing.T) {
	t.Parallel()

	mock := &mockKeyProvider{err: fmt.Errorf("wallet locked")}
	provider := NewProvider(mock, testLogger())

	did, err := provider.DID(context.Background())
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "wallet locked")
}

// mockKeyProvider implements KeyProvider for testing.
type mockKeyProvider struct {
	pubkey []byte
	err    error
	calls  int
}

func (m *mockKeyProvider) PublicKey(_ context.Context) ([]byte, error) {
	m.calls++
	return m.pubkey, m.err
}
