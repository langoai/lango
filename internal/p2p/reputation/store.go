// Package reputation tracks peer trust scores based on exchange outcomes.
package reputation

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/peerreputation"
	"go.uber.org/zap"
)

// PeerDetails holds full reputation information for a single peer.
type PeerDetails struct {
	PeerDID             string    `json:"peerDid"`
	TrustScore          float64   `json:"trustScore"`
	SuccessfulExchanges int       `json:"successfulExchanges"`
	FailedExchanges     int       `json:"failedExchanges"`
	TimeoutCount        int       `json:"timeoutCount"`
	FirstSeen           time.Time `json:"firstSeen"`
	LastInteraction     time.Time `json:"lastInteraction"`
}

// Store persists and queries peer reputation data.
type Store struct {
	client           *ent.Client
	logger           *zap.SugaredLogger
	onChangeCallback func(peerDID string, newScore float64)
}

// NewStore creates a reputation store backed by the given ent client.
func NewStore(client *ent.Client, logger *zap.SugaredLogger) *Store {
	return &Store{client: client, logger: logger}
}

// SetOnChangeCallback registers a function to be called whenever a peer's
// trust score changes. This enables reactive security measures such as
// session invalidation when scores drop below a threshold.
func (s *Store) SetOnChangeCallback(fn func(peerDID string, newScore float64)) {
	s.onChangeCallback = fn
}

// RecordSuccess increments the successful exchange count for a peer and
// recalculates the trust score.
func (s *Store) RecordSuccess(ctx context.Context, peerDID string) error {
	return s.upsert(ctx, peerDID, func(successes, failures, timeouts int) (int, int, int) {
		return successes + 1, failures, timeouts
	})
}

// RecordFailure increments the failed exchange count for a peer and
// recalculates the trust score.
func (s *Store) RecordFailure(ctx context.Context, peerDID string) error {
	return s.upsert(ctx, peerDID, func(successes, failures, timeouts int) (int, int, int) {
		return successes, failures + 1, timeouts
	})
}

// RecordTimeout increments the timeout count for a peer and recalculates the
// trust score.
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
	return &PeerDetails{
		PeerDID:             rep.PeerDid,
		TrustScore:          rep.TrustScore,
		SuccessfulExchanges: rep.SuccessfulExchanges,
		FailedExchanges:     rep.FailedExchanges,
		TimeoutCount:        rep.TimeoutCount,
		FirstSeen:           rep.FirstSeen,
		LastInteraction:     rep.LastInteraction,
	}, nil
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
		s.logger.Debugw("peer reputation created", "peerDID", peerDID, "score", score)
		if s.onChangeCallback != nil {
			s.onChangeCallback(peerDID, score)
		}
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
	s.logger.Debugw("peer reputation updated", "peerDID", peerDID, "score", score)
	if s.onChangeCallback != nil {
		s.onChangeCallback(peerDID, score)
	}
	return nil
}

// CalculateScore computes a trust score in the range [0, 1).
// Formula: successes / (successes + failures*2 + timeouts*1.5 + 1.0)
func CalculateScore(successes, failures, timeouts int) float64 {
	s := float64(successes)
	return s / (s + float64(failures)*2 + float64(timeouts)*1.5 + 1.0)
}
