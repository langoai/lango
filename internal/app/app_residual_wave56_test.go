package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/testutil"
)

func TestWave56RunLedgerModuleMetadataAndEnablement(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = false
	module := &runLedgerModule{cfg: cfg}

	require.Equal(t, "runledger", module.Name())
	require.Equal(t, []appinit.Provides{appinit.ProvidesRunLedger}, module.Provides())
	require.Equal(t, []appinit.Provides{
		appinit.ProvidesSupervisor,
		appinit.ProvidesMission,
	}, module.DependsOn())
	require.False(t, module.Enabled())

	cfg.RunLedger.Enabled = true
	require.True(t, module.Enabled())
}

func TestWave56RuntimeTrustPolicyDefaultsAndExplicitThreshold(t *testing.T) {
	t.Parallel()

	defaultPolicy := runtimeTrustPolicy(0)
	require.Equal(t, 0.3, defaultPolicy.BootstrapTrustScore)
	require.Equal(t, 0.3, defaultPolicy.MinEarnedTrustScore)
	require.Equal(t, runtimeTemporarySafetySignalLimit, defaultPolicy.MaxTemporarySafetySignals)

	explicitPolicy := runtimeTrustPolicy(0.72)
	require.Equal(t, 0.72, explicitPolicy.BootstrapTrustScore)
	require.Equal(t, 0.72, explicitPolicy.MinEarnedTrustScore)
	require.Equal(t, runtimeTemporarySafetySignalLimit, explicitPolicy.MaxTemporarySafetySignals)
}

func TestWave56RuntimePostPayTrustScoreBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	score, err := runtimePostPayTrustScore(ctx, nil, "did:lango:nil-store", 0.4)
	require.NoError(t, err)
	require.Zero(t, score)

	store := reputation.NewStore(testutil.TestEntClient(t), testutil.NopLogger())
	require.NoError(t, store.RecordSuccess(ctx, "did:lango:trusted"))

	score, err = runtimePostPayTrustScore(ctx, store, "did:lango:trusted", 0.4)
	require.NoError(t, err)
	require.Equal(t, reputation.CalculateEarnedScore(1, 0), score)

	score, err = runtimePostPayTrustScore(ctx, store, "did:lango:trusted", 0.9)
	require.NoError(t, err)
	require.Zero(t, score)
}

func TestWave56RuntimeTrustKickReasonBranches(t *testing.T) {
	t.Parallel()

	reason, shouldKick := runtimeTrustKickReason(nil)
	require.Empty(t, reason)
	require.False(t, shouldKick)

	reason, shouldKick = runtimeTrustKickReason(&reputation.TrustEntry{
		ReturningPeer: true,
		Allowed:       true,
		State:         reputation.TrustEntryStateEstablished,
	})
	require.Empty(t, reason)
	require.False(t, shouldKick)

	reason, shouldKick = runtimeTrustKickReason(&reputation.TrustEntry{
		ReturningPeer: true,
		Allowed:       false,
		State:         reputation.TrustEntryStateTemporarilyUnsafe,
	})
	require.Equal(t, "temporarily unsafe", reason)
	require.True(t, shouldKick)

	reason, shouldKick = runtimeTrustKickReason(&reputation.TrustEntry{
		ReturningPeer: true,
		Allowed:       false,
		State:         reputation.TrustEntryStateReview,
	})
	require.Equal(t, "trust entry requires review", reason)
	require.True(t, shouldKick)

	reason, shouldKick = runtimeTrustKickReason(&reputation.TrustEntry{
		ReturningPeer: true,
		Allowed:       false,
		State:         reputation.TrustEntryStateBootstrap,
	})
	require.Equal(t, "trust entry no longer allows runtime collaboration", reason)
	require.True(t, shouldKick)
}
