package app

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/langoai/lango/internal/p2p/handshake"
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
		give            string
		autoApprove     bool
		hasRepStore     bool
		wantApproved    bool
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
