// Package reputation tracks peer trust scores based on exchange outcomes.
package reputation

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/peerreputation"
	"github.com/langoai/lango/internal/eventbus"
	"go.uber.org/zap"
)

// PeerDetails holds full reputation information for a single peer.
type PeerDetails struct {
	PeerDID                string    `json:"peerDid"`
	TrustScore             float64   `json:"trustScore"`
	EarnedTrustScore       float64   `json:"earnedTrustScore"`
	SuccessfulExchanges    int       `json:"successfulExchanges"`
	FailedExchanges        int       `json:"failedExchanges"`
	DurableNegativeUnits   int       `json:"durableNegativeUnits"`
	TimeoutCount           int       `json:"timeoutCount"`
	TemporarySafetySignals int       `json:"temporarySafetySignals"`
	FirstSeen              time.Time `json:"firstSeen"`
	LastInteraction        time.Time `json:"lastInteraction"`
}

// Store persists and queries peer reputation data.
type Store struct {
	client *ent.Client
	logger *zap.SugaredLogger
	bus    *eventbus.Bus // Optional event bus for reputation change notifications.
}

// NewStore creates a reputation store backed by the given ent client.
func NewStore(client *ent.Client, logger *zap.SugaredLogger) *Store {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	return &Store{client: client, logger: logger}
}

// SetEventBus sets the optional event bus for publishing reputation change events.
func (s *Store) SetEventBus(bus *eventbus.Bus) {
	s.bus = bus
}

// publishReputationChanged publishes a ReputationChangedEvent if the bus is configured.
func (s *Store) publishReputationChanged(peerDID string, newScore float64) {
	if s.bus == nil {
		return
	}
	s.bus.Publish(eventbus.ReputationChangedEvent{
		PeerDID:  peerDID,
		NewScore: newScore,
	})
}

// RecordSuccess increments the successful exchange count for a peer and
// recalculates the trust score.
func (s *Store) RecordSuccess(ctx context.Context, peerDID string) error {
	return s.upsert(ctx, peerDID, func(successes, failures, timeouts int) (int, int, int) {
		return successes + 1, failures, timeouts
	})
}

// RecordFailure increments the durable negative exchange count for a peer and
// recalculates the composite trust score.
func (s *Store) RecordFailure(ctx context.Context, peerDID string) error {
	return s.upsert(ctx, peerDID, func(successes, failures, timeouts int) (int, int, int) {
		return successes, failures + FailurePenaltyUnits, timeouts
	})
}

// RecordTimeout increments the operational incident count for a peer and
// recalculates the composite trust score.
func (s *Store) RecordTimeout(ctx context.Context, peerDID string) error {
	return s.upsert(ctx, peerDID, func(successes, failures, timeouts int) (int, int, int) {
		return successes, failures, timeouts + 1
	})
}

// GetDetails returns full reputation details for a peer. Returns nil if the
// peer has no reputation record.
func (s *Store) GetDetails(ctx context.Context, peerDID string) (*PeerDetails, error) {
	rep, err := s.client.PeerReputation.Query().
		Where(peerreputation.PeerDid(peerDID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("query peer reputation %q: %w", peerDID, err)
	}
	return buildPeerDetails(rep), nil
}

// GetScore returns the current trust score for a peer. Returns 0.0 if the peer
// has no reputation record.
func (s *Store) GetScore(ctx context.Context, peerDID string) (float64, error) {
	rep, err := s.client.PeerReputation.Query().
		Where(peerreputation.PeerDid(peerDID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0.0, nil
		}
		return 0.0, fmt.Errorf("query peer reputation %q: %w", peerDID, err)
	}
	return rep.TrustScore, nil
}

// IsTrusted returns true if the peer's trust score meets the minimum threshold.
// New peers with no reputation record are given the benefit of the doubt and
// return true.
func (s *Store) IsTrusted(ctx context.Context, peerDID string, minScore float64) (bool, error) {
	rep, err := s.client.PeerReputation.Query().
		Where(peerreputation.PeerDid(peerDID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return true, nil // benefit of the doubt for new peers
		}
		return false, fmt.Errorf("query peer reputation %q: %w", peerDID, err)
	}
	return rep.TrustScore >= minScore, nil
}

// upsert finds or creates a peer reputation record, applies the mutator to
// adjust counters, recalculates the score, and saves.
func (s *Store) upsert(
	ctx context.Context,
	peerDID string,
	mutate func(successes, failures, timeouts int) (int, int, int),
) error {
	rep, err := s.client.PeerReputation.Query().
		Where(peerreputation.PeerDid(peerDID)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("query peer reputation %q: %w", peerDID, err)
	}

	if ent.IsNotFound(err) {
		// Create new record.
		successes, failures, timeouts := mutate(0, 0, 0)
		score := CalculateScore(successes, failures, timeouts)
		_, createErr := s.client.PeerReputation.Create().
			SetPeerDid(peerDID).
			SetSuccessfulExchanges(successes).
			SetFailedExchanges(failures).
			SetTimeoutCount(timeouts).
			SetTrustScore(score).
			Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create peer reputation %q: %w", peerDID, createErr)
		}
		s.debugw("peer reputation created", "peerDID", peerDID, "score", score)
		s.publishReputationChanged(peerDID, score)
		return nil
	}

	// Update existing record.
	successes, failures, timeouts := mutate(
		rep.SuccessfulExchanges,
		rep.FailedExchanges,
		rep.TimeoutCount,
	)
	score := CalculateScore(successes, failures, timeouts)
	_, err = s.client.PeerReputation.UpdateOne(rep).
		SetSuccessfulExchanges(successes).
		SetFailedExchanges(failures).
		SetTimeoutCount(timeouts).
		SetTrustScore(score).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update peer reputation %q: %w", peerDID, err)
	}
	s.debugw("peer reputation updated", "peerDID", peerDID, "score", score)
	s.publishReputationChanged(peerDID, score)
	return nil
}

func buildPeerDetails(rep *ent.PeerReputation) *PeerDetails {
	return &PeerDetails{
		PeerDID:                rep.PeerDid,
		TrustScore:             rep.TrustScore,
		EarnedTrustScore:       CalculateEarnedScore(rep.SuccessfulExchanges, rep.FailedExchanges),
		SuccessfulExchanges:    rep.SuccessfulExchanges,
		FailedExchanges:        rep.FailedExchanges,
		DurableNegativeUnits:   rep.FailedExchanges,
		TimeoutCount:           rep.TimeoutCount,
		TemporarySafetySignals: rep.TimeoutCount,
		FirstSeen:              rep.FirstSeen,
		LastInteraction:        rep.LastInteraction,
	}
}

func (s *Store) debugw(msg string, keysAndValues ...interface{}) {
	if s == nil || s.logger == nil {
		return
	}
	s.logger.Debugw(msg, keysAndValues...)
}

// Scoring weight constants used by CalculateScore.
const (
	FailureWeight = 2.0
	TimeoutWeight = 1.5
	BasePenalty   = 1.0
)

// CalculateScore computes a trust score in the range [0, 1).
// Formula: successes / (successes + failures*FailureWeight + timeouts*TimeoutWeight + BasePenalty)
func CalculateScore(successes, failures, timeouts int) float64 {
	s := float64(successes)
	return s / (s + float64(failures)*FailureWeight + float64(timeouts)*TimeoutWeight + BasePenalty)
}
