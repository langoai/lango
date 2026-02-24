package handshake

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestSecurityHandler(t *testing.T, maxFailures int, minTrust float64) (*SecurityEventHandler, *SessionStore) {
	t.Helper()
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	handler := NewSecurityEventHandler(store, maxFailures, minTrust, zap.NewNop().Sugar())
	return handler, store
}

func TestConsecutiveFailures_TriggerAutoInvalidation(t *testing.T) {
	handler, store := newTestSecurityHandler(t, 3, 0.3)

	sess, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	// First two failures should not invalidate.
	handler.RecordToolFailure("did:lango:peer1")
	assert.True(t, store.Validate("did:lango:peer1", sess.Token))

	handler.RecordToolFailure("did:lango:peer1")
	assert.True(t, store.Validate("did:lango:peer1", sess.Token))

	// Third failure should trigger auto-invalidation.
	handler.RecordToolFailure("did:lango:peer1")
	assert.False(t, store.Validate("did:lango:peer1", sess.Token))

	// History should record the invalidation.
	history := store.InvalidationHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, ReasonRepeatedFailures, history[0].Reason)
}

func TestSuccess_ResetsFailureCounter(t *testing.T) {
	handler, store := newTestSecurityHandler(t, 3, 0.3)

	sess, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	handler.RecordToolFailure("did:lango:peer1")
	handler.RecordToolFailure("did:lango:peer1")

	// Success resets counter.
	handler.RecordToolSuccess("did:lango:peer1")

	// Two more failures should not trigger invalidation (counter was reset).
	handler.RecordToolFailure("did:lango:peer1")
	handler.RecordToolFailure("did:lango:peer1")
	assert.True(t, store.Validate("did:lango:peer1", sess.Token))

	// Third failure after reset should trigger it.
	handler.RecordToolFailure("did:lango:peer1")
	assert.False(t, store.Validate("did:lango:peer1", sess.Token))
}

func TestReputationDrop_TriggersInvalidation(t *testing.T) {
	handler, store := newTestSecurityHandler(t, 5, 0.3)

	sess, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	// Score above threshold should not invalidate.
	handler.OnReputationChange("did:lango:peer1", 0.5)
	assert.True(t, store.Validate("did:lango:peer1", sess.Token))

	// Score below threshold should invalidate.
	handler.OnReputationChange("did:lango:peer1", 0.2)
	assert.False(t, store.Validate("did:lango:peer1", sess.Token))

	history := store.InvalidationHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, ReasonReputationDrop, history[0].Reason)
}

func TestReputationAtThreshold_NoInvalidation(t *testing.T) {
	handler, store := newTestSecurityHandler(t, 5, 0.3)

	sess, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	// Score exactly at threshold should not invalidate.
	handler.OnReputationChange("did:lango:peer1", 0.3)
	assert.True(t, store.Validate("did:lango:peer1", sess.Token))
}

func TestDefaultMaxFailures(t *testing.T) {
	store, err := NewSessionStore(24 * time.Hour)
	require.NoError(t, err)

	// Pass 0 for maxFailures; should default to 5.
	handler := NewSecurityEventHandler(store, 0, 0.3, zap.NewNop().Sugar())

	sess, err := store.Create("did:lango:peer1", false)
	require.NoError(t, err)

	// 4 failures should not trigger invalidation.
	for i := 0; i < 4; i++ {
		handler.RecordToolFailure("did:lango:peer1")
	}
	assert.True(t, store.Validate("did:lango:peer1", sess.Token))

	// 5th failure should trigger it.
	handler.RecordToolFailure("did:lango:peer1")
	assert.False(t, store.Validate("did:lango:peer1", sess.Token))
}
