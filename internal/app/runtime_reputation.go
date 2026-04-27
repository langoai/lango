package app

import (
	"context"

	"github.com/langoai/lango/internal/p2p/reputation"
)

const runtimeTemporarySafetySignalLimit = 1

func runtimeTrustPolicy(minScore float64) reputation.TrustEntryPolicy {
	if minScore <= 0 {
		minScore = 0.3
	}
	return reputation.TrustEntryPolicy{
		BootstrapTrustScore:       minScore,
		MinEarnedTrustScore:       minScore,
		MaxTemporarySafetySignals: runtimeTemporarySafetySignalLimit,
	}
}

func runtimeTrustEntry(ctx context.Context, store *reputation.Store, peerDID string, minScore float64) (*reputation.TrustEntry, error) {
	if store == nil {
		return nil, nil
	}
	return store.GetTrustEntry(ctx, peerDID, runtimeTrustPolicy(minScore))
}

func autoApproveKnownPeer(ctx context.Context, store *reputation.Store, peerDID string, minScore float64) (bool, error) {
	entry, err := runtimeTrustEntry(ctx, store, peerDID, minScore)
	if err != nil || entry == nil {
		return false, err
	}
	return entry.ReturningPeer &&
		entry.State == reputation.TrustEntryStateEstablished &&
		entry.Allowed &&
		!entry.RequiresApproval, nil
}

func runtimeEconomyTrustScore(ctx context.Context, store *reputation.Store, peerDID string, minScore float64) (float64, error) {
	entry, err := runtimeTrustEntry(ctx, store, peerDID, minScore)
	if err != nil || entry == nil {
		return 0, err
	}
	if !entry.ReturningPeer {
		return entry.EffectiveTrustScore, nil
	}
	return entry.EarnedTrustScore, nil
}

func runtimePostPayTrustScore(ctx context.Context, store *reputation.Store, peerDID string, minScore float64) (float64, error) {
	entry, err := runtimeTrustEntry(ctx, store, peerDID, minScore)
	if err != nil || entry == nil {
		return 0, err
	}
	if entry.State != reputation.TrustEntryStateEstablished || !entry.Allowed || entry.RequiresApproval {
		return 0, nil
	}
	return entry.EarnedTrustScore, nil
}

func runtimeTrustKickReason(entry *reputation.TrustEntry) (string, bool) {
	if entry == nil || !entry.ReturningPeer || entry.Allowed {
		return "", false
	}
	switch entry.State {
	case reputation.TrustEntryStateTemporarilyUnsafe:
		return "temporarily unsafe", true
	case reputation.TrustEntryStateReview:
		return "trust entry requires review", true
	default:
		return "trust entry no longer allows runtime collaboration", true
	}
}
