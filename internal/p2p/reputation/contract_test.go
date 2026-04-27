package reputation_test

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/testutil"
)

func TestPeerDetails_CanonicalSignalsSeparateDurableAndTemporary(t *testing.T) {
	t.Parallel()

	details := &reputation.PeerDetails{
		PeerDID:             "did:lango:peer",
		TrustScore:          reputation.CalculateScore(4, 2, 3),
		SuccessfulExchanges: 4,
		FailedExchanges:     2,
		TimeoutCount:        3,
	}

	signals := details.CanonicalSignals()

	assert.Equal(t, 2, signals.DurableNegativeUnits)
	assert.Equal(t, 3, signals.TemporarySafetySignals)
	assert.True(t, signals.ReturningPeer)
	assert.InDelta(t, reputation.CalculateEarnedScore(4, 2), signals.EarnedTrustScore, 1e-9)
	assert.InDelta(t, details.TrustScore, signals.CompatibilityScore, 1e-9)
}

func TestStore_RecordAdjudicatedFailureCarriesStrongerDurablePenalty(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := reputation.NewStore(client, testutil.NopLogger())
	ctx := context.Background()

	require.NoError(t, store.RecordSuccess(ctx, "did:lango:standard"))
	require.NoError(t, store.RecordFailure(ctx, "did:lango:standard"))
	require.NoError(t, store.RecordSuccess(ctx, "did:lango:adjudicated"))
	require.NoError(t, store.RecordAdjudicatedFailure(ctx, "did:lango:adjudicated"))

	standard, err := store.GetDetails(ctx, "did:lango:standard")
	require.NoError(t, err)
	require.NotNil(t, standard)

	adjudicated, err := store.GetDetails(ctx, "did:lango:adjudicated")
	require.NoError(t, err)
	require.NotNil(t, adjudicated)

	assert.Equal(t, reputation.FailurePenaltyUnits, standard.FailedExchanges)
	assert.Equal(t, reputation.AdjudicatedFailurePenaltyUnits, adjudicated.FailedExchanges)
	assert.Less(t, adjudicated.EarnedTrustScore, standard.EarnedTrustScore)
}

func TestStore_GetTrustEntry_BootstrapAndReturningRemainDistinct(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := reputation.NewStore(client, testutil.NopLogger())
	ctx := context.Background()
	policy := reputation.TrustEntryPolicy{
		BootstrapTrustScore:       0.65,
		MinEarnedTrustScore:       0.50,
		MaxTemporarySafetySignals: 2,
	}

	entry, err := store.GetTrustEntry(ctx, "did:lango:new-peer", policy)
	require.NoError(t, err)

	assert.Equal(t, reputation.TrustEntryStateBootstrap, entry.State)
	assert.True(t, entry.Allowed)
	assert.True(t, entry.BootstrapTrust)
	assert.False(t, entry.ReturningPeer)
	assert.InDelta(t, 0.65, entry.EffectiveTrustScore, 1e-9)
	assert.InDelta(t, 0.0, entry.EarnedTrustScore, 1e-9)

	require.NoError(t, store.RecordSuccess(ctx, "did:lango:new-peer"))

	entry, err = store.GetTrustEntry(ctx, "did:lango:new-peer", policy)
	require.NoError(t, err)

	assert.Equal(t, reputation.TrustEntryStateEstablished, entry.State)
	assert.True(t, entry.Allowed)
	assert.False(t, entry.BootstrapTrust)
	assert.True(t, entry.ReturningPeer)
	assert.Greater(t, entry.EarnedTrustScore, 0.0)
}

func TestStore_GetTrustEntry_LowEarnedTrustKeepsOwnerRootSeparate(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := reputation.NewStore(client, testutil.NopLogger())
	ctx := context.Background()
	require.NoError(t, store.RecordFailure(ctx, "did:lango:review-peer"))

	basePolicy := reputation.TrustEntryPolicy{
		BootstrapTrustScore:       0.60,
		MinEarnedTrustScore:       0.50,
		MaxTemporarySafetySignals: 2,
	}

	withoutRoot, err := store.GetTrustEntry(ctx, "did:lango:review-peer", basePolicy)
	require.NoError(t, err)

	withRoot, err := store.GetTrustEntry(ctx, "did:lango:review-peer", reputation.TrustEntryPolicy{
		OwnerRootTrusted:          true,
		BootstrapTrustScore:       basePolicy.BootstrapTrustScore,
		MinEarnedTrustScore:       basePolicy.MinEarnedTrustScore,
		MaxTemporarySafetySignals: basePolicy.MaxTemporarySafetySignals,
	})
	require.NoError(t, err)

	assert.Equal(t, reputation.TrustEntryStateReview, withoutRoot.State)
	assert.True(t, withoutRoot.RequiresApproval)
	assert.False(t, withoutRoot.OwnerContinuityFloor)
	assert.InDelta(t, withoutRoot.EarnedTrustScore, withoutRoot.EffectiveTrustScore, 1e-9)

	assert.Equal(t, reputation.TrustEntryStateReview, withRoot.State)
	assert.True(t, withRoot.RequiresApproval)
	assert.True(t, withRoot.OwnerContinuityFloor)
	assert.InDelta(t, withoutRoot.EarnedTrustScore, withRoot.EarnedTrustScore, 1e-9)
	assert.InDelta(t, basePolicy.BootstrapTrustScore, withRoot.EffectiveTrustScore, 1e-9)
}

func TestStore_GetTrustEntry_TemporarySafetySignalsStayOutOfDurableReputation(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := reputation.NewStore(client, testutil.NopLogger())
	ctx := context.Background()
	peerDID := "did:lango:temporarily-unsafe"

	require.NoError(t, store.RecordSuccess(ctx, peerDID))
	require.NoError(t, store.RecordTimeout(ctx, peerDID))

	details, err := store.GetDetails(ctx, peerDID)
	require.NoError(t, err)
	require.NotNil(t, details)

	assert.Equal(t, 0, details.DurableNegativeUnits)
	assert.Equal(t, 1, details.TemporarySafetySignals)
	assert.InDelta(t, reputation.CalculateEarnedScore(1, 0), details.EarnedTrustScore, 1e-9)
	assert.Greater(t, details.EarnedTrustScore, details.TrustScore)

	entry, err := store.GetTrustEntry(ctx, peerDID, reputation.TrustEntryPolicy{
		BootstrapTrustScore:       0.60,
		MinEarnedTrustScore:       0.50,
		MaxTemporarySafetySignals: 1,
	})
	require.NoError(t, err)

	assert.Equal(t, reputation.TrustEntryStateTemporarilyUnsafe, entry.State)
	assert.False(t, entry.Allowed)
	assert.False(t, entry.RequiresApproval)
	assert.Equal(t, 0, entry.DurableNegativeUnits)
	assert.Equal(t, 1, entry.TemporarySafetySignals)
}

func TestEvaluateTrustEntry_ClampsNaNBootstrapScore(t *testing.T) {
	t.Parallel()

	entry := reputation.EvaluateTrustEntry("did:lango:new-peer", nil, reputation.TrustEntryPolicy{
		BootstrapTrustScore:       math.NaN(),
		MinEarnedTrustScore:       0.5,
		MaxTemporarySafetySignals: 1,
	})

	assert.False(t, math.IsNaN(entry.EffectiveTrustScore))
	assert.InDelta(t, 0.0, entry.EffectiveTrustScore, 1e-9)
}
