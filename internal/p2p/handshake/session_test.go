package handshake

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidate_SessionBecomesInvalid(t *testing.T) {
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	sess, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)
	require.NotEmpty(t, sess.Token)

	// Session should be valid before invalidation.
	assert.True(t, store.Validate("did:lango:peer1", sess.Token))

	// Invalidate the session.
	store.Invalidate("did:lango:peer1", ReasonManualRevoke)

	// Session should no longer validate.
	assert.False(t, store.Validate("did:lango:peer1", sess.Token))

	// Get should return nil for invalidated session.
	assert.Nil(t, store.Get("did:lango:peer1"))
}

func TestInvalidateAll_AllSessionsInvalidated(t *testing.T) {
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	_, err = store.Create("did:lango:peer1", false)
	require.NoError(t, err)
	_, err = store.Create("did:lango:peer2", true)
	require.NoError(t, err)
	_, err = store.Create("did:lango:peer3", false)
	require.NoError(t, err)

	assert.Len(t, store.ActiveSessions(), 3)

	store.InvalidateAll(ReasonSecurityEvent)

	assert.Empty(t, store.ActiveSessions())

	// History should contain all three invalidations.
	history := store.InvalidationHistory()
	assert.Len(t, history, 3)
	for _, rec := range history {
		assert.Equal(t, ReasonSecurityEvent, rec.Reason)
	}
}

func TestInvalidateByCondition_SelectiveInvalidation(t *testing.T) {
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	sess1, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)
	sess2, err := store.Create("did:lango:peer2", true)
	require.NoError(t, err)

	// Invalidate only non-ZK-verified sessions.
	store.InvalidateByCondition(ReasonSecurityEvent, func(s *Session) bool {
		return !s.ZKVerified
	})

	// peer1 (non-ZK) should be invalidated; peer2 (ZK) should remain.
	assert.False(t, store.Validate("did:lango:peer1", sess1.Token))
	assert.True(t, store.Validate("did:lango:peer2", sess2.Token))

	active := store.ActiveSessions()
	assert.Len(t, active, 1)
	assert.Equal(t, "did:lango:peer2", active[0].PeerDID)
}

func TestValidate_ReturnsFalseForInvalidated(t *testing.T) {
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	sess, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	assert.True(t, store.Validate("did:lango:peer1", sess.Token))

	store.Invalidate("did:lango:peer1", ReasonLogout)

	assert.False(t, store.Validate("did:lango:peer1", sess.Token))
}

func TestInvalidationHistory_ReturnsRecords(t *testing.T) {
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	_, err = store.Create("did:lango:peer1", false)
	require.NoError(t, err)
	_, err = store.Create("did:lango:peer2", false)
	require.NoError(t, err)

	assert.Empty(t, store.InvalidationHistory())

	store.Invalidate("did:lango:peer1", ReasonReputationDrop)
	store.Invalidate("did:lango:peer2", ReasonRepeatedFailures)

	history := store.InvalidationHistory()
	assert.Len(t, history, 2)

	assert.Equal(t, "did:lango:peer1", history[0].PeerDID)
	assert.Equal(t, ReasonReputationDrop, history[0].Reason)
	assert.False(t, history[0].InvalidatedAt.IsZero())

	assert.Equal(t, "did:lango:peer2", history[1].PeerDID)
	assert.Equal(t, ReasonRepeatedFailures, history[1].Reason)
}

func TestInvalidationCallback_FiredOnInvalidate(t *testing.T) {
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	var callbackDID string
	var callbackReason InvalidationReason
	store.SetInvalidationCallback(func(peerDID string, reason InvalidationReason) {
		callbackDID = peerDID
		callbackReason = reason
	})

	_, err = store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	store.Invalidate("did:lango:peer1", ReasonManualRevoke)

	assert.Equal(t, "did:lango:peer1", callbackDID)
	assert.Equal(t, ReasonManualRevoke, callbackReason)
}

func TestInvalidateNonExistent_StillRecordsHistory(t *testing.T) {
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	// Invalidating a non-existent session should still record history.
	store.Invalidate("did:lango:unknown", ReasonSecurityEvent)

	history := store.InvalidationHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "did:lango:unknown", history[0].PeerDID)
}

func TestCleanup_RemovesInvalidatedSessions(t *testing.T) {
	store, err := NewSessionStore(1 * time.Millisecond)
	require.NoError(t, err)

	_, err = store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	// Wait for expiry.
	time.Sleep(5 * time.Millisecond)

	removed := store.Cleanup()
	assert.Equal(t, 1, removed)
}
