package identity

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	assert.Equal(t, "did:lango:", DIDPrefix)
}

func TestDIDFromPublicKey_Valid(t *testing.T) {
	pubkey := generateTestPubkey(t)

	did, err := DIDFromPublicKey(pubkey)
	require.NoError(t, err)
	require.NotNil(t, did)

	assert.True(t, strings.HasPrefix(did.ID, DIDPrefix))
	assert.Equal(t, pubkey, did.PublicKey)
	assert.NotEmpty(t, did.PeerID)

	// Verify the hex encoding in the DID string.
	hexPart := strings.TrimPrefix(did.ID, DIDPrefix)
	decoded, err := hex.DecodeString(hexPart)
	require.NoError(t, err)
	assert.Equal(t, pubkey, decoded)
}

func TestDIDFromPublicKey_EmptyKey(t *testing.T) {
	did, err := DIDFromPublicKey(nil)
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "empty public key")

	did, err = DIDFromPublicKey([]byte{})
	assert.Error(t, err)
	assert.Nil(t, did)
}

func TestParseDID_Valid_Roundtrip(t *testing.T) {
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
	did, err := ParseDID("did:other:abc123")
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "invalid DID scheme")
}

func TestParseDID_EmptyKey(t *testing.T) {
	did, err := ParseDID("did:lango:")
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "empty public key")
}

func TestParseDID_InvalidHex(t *testing.T) {
	did, err := ParseDID("did:lango:ZZZZ_not_hex")
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "decode hex")
}

func TestVerifyDID_Matching(t *testing.T) {
	pubkey := generateTestPubkey(t)
	did, err := DIDFromPublicKey(pubkey)
	require.NoError(t, err)

	provider := NewProvider(&mockWalletProvider{pubkey: pubkey}, testLogger())
	err = provider.VerifyDID(did, did.PeerID)
	assert.NoError(t, err)
}

func TestVerifyDID_Mismatched(t *testing.T) {
	pubkey := generateTestPubkey(t)
	did, err := DIDFromPublicKey(pubkey)
	require.NoError(t, err)

	// Generate a different peer ID.
	otherPubkey := generateTestPubkey(t)
	otherDID, err := DIDFromPublicKey(otherPubkey)
	require.NoError(t, err)

	provider := NewProvider(&mockWalletProvider{pubkey: pubkey}, testLogger())
	err = provider.VerifyDID(did, otherDID.PeerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "peer ID mismatch")
}

func TestVerifyDID_NilDID(t *testing.T) {
	provider := NewProvider(&mockWalletProvider{}, testLogger())
	err := provider.VerifyDID(nil, peer.ID("somepeerid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil DID")
}

func TestWalletDIDProvider_DID_Caching(t *testing.T) {
	pubkey := generateTestPubkey(t)
	mock := &mockWalletProvider{pubkey: pubkey}
	provider := NewProvider(mock, testLogger())

	did1, err := provider.DID(context.Background())
	require.NoError(t, err)

	did2, err := provider.DID(context.Background())
	require.NoError(t, err)

	assert.Same(t, did1, did2, "second call should return cached DID")
	assert.Equal(t, 1, mock.calls, "PublicKey should only be called once due to caching")
}

func TestWalletDIDProvider_DID_WalletError(t *testing.T) {
	mock := &mockWalletProvider{err: fmt.Errorf("wallet locked")}
	provider := NewProvider(mock, testLogger())

	did, err := provider.DID(context.Background())
	assert.Error(t, err)
	assert.Nil(t, did)
	assert.Contains(t, err.Error(), "wallet locked")
}

// mockWalletProvider implements wallet.WalletProvider for testing.
type mockWalletProvider struct {
	pubkey []byte
	err    error
	calls  int
}

func (m *mockWalletProvider) PublicKey(_ context.Context) ([]byte, error) {
	m.calls++
	return m.pubkey, m.err
}

func (m *mockWalletProvider) SignMessage(_ context.Context, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockWalletProvider) SignTransaction(_ context.Context, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockWalletProvider) Address(_ context.Context) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (m *mockWalletProvider) Balance(_ context.Context) (*big.Int, error) {
	return nil, fmt.Errorf("not implemented")
}
