package app

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/paygate"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// WU-E2 Test 1: NonceCache lifecycle (start → record → replay → TTL expire → stop)
// ---------------------------------------------------------------------------

func TestNonceCacheLifecycle(t *testing.T) {
	t.Parallel()

	ttl := 150 * time.Millisecond
	nc := handshake.NewNonceCache(ttl)
	nc.Start()
	defer nc.Stop()

	// Generate a valid 32-byte nonce.
	nonce := make([]byte, handshake.NonceSize)
	_, err := rand.Read(nonce)
	require.NoError(t, err)

	// First use: should be accepted (new nonce).
	ok := nc.CheckAndRecord(nonce)
	assert.True(t, ok, "first occurrence of nonce should return true")

	// Replay: same nonce should be rejected.
	ok = nc.CheckAndRecord(nonce)
	assert.False(t, ok, "replay of same nonce should return false")

	// Wait for TTL expiry + cleanup cycle (ticker fires at ttl/2).
	time.Sleep(ttl + ttl/2 + 50*time.Millisecond)

	// After expiry + cleanup the nonce should be accepted again.
	ok = nc.CheckAndRecord(nonce)
	assert.True(t, ok, "nonce should be accepted after TTL expiry")
}

func TestNonceCacheLifecycle_InvalidSize(t *testing.T) {
	t.Parallel()

	nc := handshake.NewNonceCache(time.Second)
	nc.Start()
	defer nc.Stop()

	// Nonces that are not exactly 32 bytes should be rejected.
	short := make([]byte, 16)
	assert.False(t, nc.CheckAndRecord(short), "short nonce should be rejected")

	long := make([]byte, 64)
	assert.False(t, nc.CheckAndRecord(long), "oversized nonce should be rejected")

	assert.False(t, nc.CheckAndRecord(nil), "nil nonce should be rejected")
}

func TestNonceCacheLifecycle_DistinctNonces(t *testing.T) {
	t.Parallel()

	nc := handshake.NewNonceCache(5 * time.Second)
	nc.Start()
	defer nc.Stop()

	nonce1 := make([]byte, handshake.NonceSize)
	nonce2 := make([]byte, handshake.NonceSize)
	_, _ = rand.Read(nonce1)
	_, _ = rand.Read(nonce2)

	assert.True(t, nc.CheckAndRecord(nonce1), "nonce1 first use should succeed")
	assert.True(t, nc.CheckAndRecord(nonce2), "nonce2 first use should succeed")
	assert.False(t, nc.CheckAndRecord(nonce1), "nonce1 replay should fail")
	assert.False(t, nc.CheckAndRecord(nonce2), "nonce2 replay should fail")
}

// ---------------------------------------------------------------------------
// WU-E2 Test 2: Default-deny approval function pattern
// ---------------------------------------------------------------------------

func TestApprovalFnDefaultDeny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		autoApprove  bool
		hasRepStore  bool
		wantApproved bool
	}{
		{
			give:         "auto-approve off, no rep store → deny",
			autoApprove:  false,
			hasRepStore:  false,
			wantApproved: false,
		},
		{
			give:         "auto-approve on, no rep store → deny",
			autoApprove:  true,
			hasRepStore:  false,
			wantApproved: false,
		},
		{
			give:         "auto-approve off, has rep store → deny",
			autoApprove:  false,
			hasRepStore:  true,
			wantApproved: false,
		},
		{
			give:         "auto-approve on, has rep store → approve",
			autoApprove:  true,
			hasRepStore:  true,
			wantApproved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			// Simulate the closure pattern from initP2P (wiring_p2p.go:110-123).
			// We capture a "repStore" that is back-filled later; the approval
			// function checks two conditions: autoApprove flag AND non-nil
			// reputation store.
			type fakeRepStore struct{}
			var repStoreRef *fakeRepStore
			if tt.hasRepStore {
				repStoreRef = &fakeRepStore{}
			}

			approvalFn := func(_ context.Context, _ *handshake.PendingHandshake) (bool, error) {
				if tt.autoApprove && repStoreRef != nil {
					// In the real code this queries reputation score; simulate
					// a peer with score above threshold.
					return true, nil
				}
				return false, nil
			}

			approved, err := approvalFn(context.Background(), &handshake.PendingHandshake{
				PeerDID: "did:example:peer1",
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantApproved, approved)
		})
	}
}

func TestApprovalFnDenyBelowMinScore(t *testing.T) {
	t.Parallel()

	// Simulate the full approval pattern with a reputation score check.
	minScore := 0.3
	peerScore := 0.1 // below threshold

	approvalFn := func(_ context.Context, _ *handshake.PendingHandshake) (bool, error) {
		// autoApprove = true, repStore = present
		return peerScore >= minScore, nil
	}

	approved, err := approvalFn(context.Background(), &handshake.PendingHandshake{
		PeerDID: "did:example:low-rep",
	})
	require.NoError(t, err)
	assert.False(t, approved, "peer below min trust score should be denied")
}

func TestApprovalFnApproveAboveMinScore(t *testing.T) {
	t.Parallel()

	minScore := 0.3
	peerScore := 0.85

	approvalFn := func(_ context.Context, _ *handshake.PendingHandshake) (bool, error) {
		return peerScore >= minScore, nil
	}

	approved, err := approvalFn(context.Background(), &handshake.PendingHandshake{
		PeerDID: "did:example:high-rep",
	})
	require.NoError(t, err)
	assert.True(t, approved, "peer above min trust score should be approved")
}

func TestAutoApproveKnownPeer_UsesTrustEntrySemantics(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := testutil.TestEntClient(t)
	store := reputation.NewStore(client, testLog())

	approved, err := autoApproveKnownPeer(ctx, store, "did:example:new", 0.3)
	require.NoError(t, err)
	assert.False(t, approved, "bootstrap peers should not be auto-approved as known peers")

	require.NoError(t, store.RecordSuccess(ctx, "did:example:established"))
	approved, err = autoApproveKnownPeer(ctx, store, "did:example:established", 0.3)
	require.NoError(t, err)
	assert.True(t, approved, "established returning peers should be auto-approved")

	require.NoError(t, store.RecordSuccess(ctx, "did:example:unsafe"))
	require.NoError(t, store.RecordOperationalIncident(ctx, "did:example:unsafe"))
	approved, err = autoApproveKnownPeer(ctx, store, "did:example:unsafe", 0.3)
	require.NoError(t, err)
	assert.False(t, approved, "temporarily unsafe peers should not be auto-approved")
}

type wiringP2PWallet struct {
	signature []byte
	publicKey []byte
	signErr   error
	pubErr    error
}

func (w *wiringP2PWallet) Address(context.Context) (string, error) {
	return "0x0000000000000000000000000000000000000001", nil
}

func (w *wiringP2PWallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w *wiringP2PWallet) SignTransaction(context.Context, []byte) ([]byte, error) {
	return nil, nil
}

func (w *wiringP2PWallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return w.signature, w.signErr
}

func (w *wiringP2PWallet) PublicKey(context.Context) ([]byte, error) {
	return w.publicKey, w.pubErr
}

func TestWalletHandshakeSigner_DelegatesWalletAndDerivesDID(t *testing.T) {
	t.Parallel()

	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	pub := ethcrypto.CompressPubkey(&key.PublicKey)
	wallet := &wiringP2PWallet{
		signature: []byte("signed-message"),
		publicKey: pub,
	}
	signer := &walletHandshakeSigner{wp: wallet}

	signature, err := signer.SignMessage(context.Background(), []byte("challenge"))
	require.NoError(t, err)
	assert.Equal(t, []byte("signed-message"), signature)

	gotPub, err := signer.PublicKey(context.Background())
	require.NoError(t, err)
	assert.Equal(t, pub, gotPub)
	assert.Equal(t, security.AlgorithmSecp256k1Keccak256, signer.Algorithm())

	did, err := signer.DID(context.Background())
	require.NoError(t, err)
	expectedDID, err := identity.DIDFromPublicKey(pub)
	require.NoError(t, err)
	assert.Equal(t, expectedDID.ID, did)
}

func TestWalletHandshakeSigner_DIDReturnsWalletPublicKeyError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("public key unavailable")
	signer := &walletHandshakeSigner{wp: &wiringP2PWallet{pubErr: wantErr}}

	did, err := signer.DID(context.Background())
	require.ErrorIs(t, err, wantErr)
	assert.Empty(t, did)
}

func TestInitP2P_SkipsDisabledOrMissingWalletWithoutNetwork(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "p2p-keys")
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}
	cfg.P2P.Enabled = false
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	assert.NoDirExists(t, cfg.P2P.KeyDir)

	cfg.P2P.Enabled = true
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	assert.NoDirExists(t, cfg.P2P.KeyDir)
}

func TestPayGateAdapter_CheckMapsPaymentRequiredQuote(t *testing.T) {
	t.Parallel()

	gate := paygate.New(paygate.Config{
		PricingFn: func(toolName string) (string, bool) {
			if toolName == "free_tool" {
				return "", true
			}
			return "1.25", false
		},
		LocalAddr: "0x00000000000000000000000000000000000000aa",
		ChainID:   8453,
		USDCAddr:  ethcommon.HexToAddress("0x00000000000000000000000000000000000000bb"),
		Logger:    testLog(),
	})
	adapter := &payGateAdapter{gate: gate}

	free, err := adapter.Check("did:lango:peer", "free_tool", nil)
	require.NoError(t, err)
	assert.Equal(t, p2pproto.PayGateResult{Status: string(paygate.StatusFree)}, free)

	paid, err := adapter.Check("did:lango:peer", "paid_tool", map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, string(paygate.StatusPaymentRequired), paid.Status)
	require.NotNil(t, paid.PriceQuote)
	assert.Equal(t, "paid_tool", paid.PriceQuote["toolName"])
	assert.Equal(t, "1.25", paid.PriceQuote["price"])
	assert.Equal(t, int64(8453), paid.PriceQuote["chainId"])
	assert.Equal(t, false, paid.PriceQuote["isFree"])
}

func TestInitZKP_ReturnsNilWhenHandshakeAndAttestationDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.ZKHandshake = false
	cfg.P2P.ZKAttestation = false

	assert.Nil(t, initZKP(cfg))
}
