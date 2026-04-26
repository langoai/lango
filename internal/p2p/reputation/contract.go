package reputation

import "context"

const (
	// FailurePenaltyUnits is the durable penalty applied for a standard
	// negative collaboration outcome.
	FailurePenaltyUnits = 1
	// AdjudicatedFailurePenaltyUnits is the stronger durable penalty applied
	// after a reviewed or adjudicated negative outcome.
	AdjudicatedFailurePenaltyUnits = 2
)

// TrustBasis describes which canonical trust basis is currently carrying a peer.
type TrustBasis string

const (
	TrustBasisBootstrap TrustBasis = "bootstrap"
	TrustBasisOwnerRoot TrustBasis = "owner_root"
	TrustBasisEarned    TrustBasis = "earned"
)

// TrustEntryState describes the canonical trust-entry state without collapsing
// every decision into one scalar score.
type TrustEntryState string

const (
	TrustEntryStateBootstrap         TrustEntryState = "bootstrap"
	TrustEntryStateEstablished       TrustEntryState = "established"
	TrustEntryStateReview            TrustEntryState = "review"
	TrustEntryStateTemporarilyUnsafe TrustEntryState = "temporarily_unsafe"
)

// CanonicalSignals exposes the separated V2 reputation signals derived from a
// peer record.
type CanonicalSignals struct {
	DurableNegativeUnits   int     `json:"durableNegativeUnits"`
	TemporarySafetySignals int     `json:"temporarySafetySignals"`
	ReturningPeer          bool    `json:"returningPeer"`
	EarnedTrustScore       float64 `json:"earnedTrustScore"`
	CompatibilityScore     float64 `json:"compatibilityScore"`
}

// TrustEntryPolicy provides caller-owned thresholds for canonical trust entry.
// The store reports the separated signals and evaluates them against this
// narrow policy surface without pushing runtime-specific actions into the
// reputation layer.
type TrustEntryPolicy struct {
	OwnerRootTrusted          bool
	BootstrapTrustScore       float64
	MinEarnedTrustScore       float64
	MaxTemporarySafetySignals int
}

// TrustEntry captures the V2 canonical contract for trust entry.
type TrustEntry struct {
	PeerDID                  string          `json:"peerDid"`
	Basis                    TrustBasis      `json:"basis"`
	State                    TrustEntryState `json:"state"`
	Allowed                  bool            `json:"allowed"`
	RequiresApproval         bool            `json:"requiresApproval"`
	OwnerRootTrusted         bool            `json:"ownerRootTrusted"`
	OwnerContinuityFloor     bool            `json:"ownerContinuityFloor"`
	BootstrapTrust           bool            `json:"bootstrapTrust"`
	BootstrapTrusted         bool            `json:"bootstrapTrusted"`
	ReturningPeer            bool            `json:"returningPeer"`
	HasEarnedReputation      bool            `json:"hasEarnedReputation"`
	EarnedTrustScore         float64         `json:"earnedTrustScore"`
	EarnedScore              float64         `json:"earnedScore"`
	CompatibilityScore       float64         `json:"compatibilityScore"`
	CompositeScore           float64         `json:"compositeScore"`
	EffectiveTrustScore      float64         `json:"effectiveTrustScore"`
	EffectiveScore           float64         `json:"effectiveScore"`
	DurableNegativeUnits     int             `json:"durableNegativeUnits"`
	DurableNegativeCount     int             `json:"durableNegativeCount"`
	TemporarySafetySignals   int             `json:"temporarySafetySignals"`
	OperationalIncidentCount int             `json:"operationalIncidentCount"`
	TemporarilyUnsafe        bool            `json:"temporarilyUnsafe"`
}

// CalculateEarnedScore returns the score derived from actual collaboration
// history only. Temporary operational incidents are intentionally excluded.
func CalculateEarnedScore(successes, failures int) float64 {
	return CalculateScore(successes, failures, 0)
}

// CanonicalSignals returns the separated V2 reputation signals for a peer.
func (d *PeerDetails) CanonicalSignals() CanonicalSignals {
	if d == nil {
		return CanonicalSignals{}
	}

	durableUnits := d.DurableNegativeUnits
	if durableUnits == 0 && d.FailedExchanges > 0 {
		durableUnits = d.FailedExchanges
	}

	temporarySignals := d.TemporarySafetySignals
	if temporarySignals == 0 && d.TimeoutCount > 0 {
		temporarySignals = d.TimeoutCount
	}

	earnedTrustScore := d.EarnedTrustScore
	if earnedTrustScore == 0 && (d.SuccessfulExchanges > 0 || d.FailedExchanges > 0) {
		earnedTrustScore = CalculateEarnedScore(d.SuccessfulExchanges, d.FailedExchanges)
	}

	returningPeer := d.SuccessfulExchanges > 0 ||
		d.FailedExchanges > 0 ||
		d.TimeoutCount > 0 ||
		!d.FirstSeen.IsZero() ||
		!d.LastInteraction.IsZero()

	return CanonicalSignals{
		DurableNegativeUnits:   durableUnits,
		TemporarySafetySignals: temporarySignals,
		ReturningPeer:          returningPeer,
		EarnedTrustScore:       earnedTrustScore,
		CompatibilityScore:     d.TrustScore,
	}
}

// EvaluateTrustEntry separates owner continuity, earned reputation, durable
// negatives, and temporary operational safety signals into one canonical view.
func EvaluateTrustEntry(peerDID string, details *PeerDetails, policy TrustEntryPolicy) TrustEntry {
	signals := details.CanonicalSignals()
	hasEarnedReputation := details != nil && (details.SuccessfulExchanges > 0 || details.FailedExchanges > 0)

	entry := TrustEntry{
		PeerDID:                  peerDID,
		OwnerRootTrusted:         policy.OwnerRootTrusted,
		BootstrapTrust:           !signals.ReturningPeer,
		BootstrapTrusted:         !signals.ReturningPeer,
		ReturningPeer:            signals.ReturningPeer,
		HasEarnedReputation:      hasEarnedReputation,
		EarnedTrustScore:         signals.EarnedTrustScore,
		EarnedScore:              signals.EarnedTrustScore,
		CompatibilityScore:       signals.CompatibilityScore,
		CompositeScore:           signals.CompatibilityScore,
		DurableNegativeUnits:     signals.DurableNegativeUnits,
		DurableNegativeCount:     signals.DurableNegativeUnits,
		TemporarySafetySignals:   signals.TemporarySafetySignals,
		OperationalIncidentCount: signals.TemporarySafetySignals,
		TemporarilyUnsafe:        false,
	}

	if !signals.ReturningPeer {
		entry.State = TrustEntryStateBootstrap
		entry.Allowed = true
		entry.Basis = TrustBasisBootstrap
		entry.EffectiveTrustScore = clampScore(policy.BootstrapTrustScore)
		entry.EffectiveScore = entry.EffectiveTrustScore
		return entry
	}

	entry.EffectiveTrustScore = entry.EarnedTrustScore
	entry.EffectiveScore = entry.EffectiveTrustScore

	if policy.OwnerRootTrusted && entry.EffectiveTrustScore < policy.BootstrapTrustScore {
		entry.OwnerContinuityFloor = true
		entry.EffectiveTrustScore = clampScore(policy.BootstrapTrustScore)
		entry.EffectiveScore = entry.EffectiveTrustScore
		entry.Basis = TrustBasisOwnerRoot
	} else {
		entry.Basis = TrustBasisEarned
	}

	if policy.MaxTemporarySafetySignals > 0 &&
		entry.TemporarySafetySignals >= policy.MaxTemporarySafetySignals {
		entry.State = TrustEntryStateTemporarilyUnsafe
		entry.TemporarilyUnsafe = true
		entry.Allowed = false
		return entry
	}

	if entry.EarnedTrustScore >= clampScore(policy.MinEarnedTrustScore) {
		entry.State = TrustEntryStateEstablished
		entry.Allowed = true
		return entry
	}

	entry.State = TrustEntryStateReview
	entry.RequiresApproval = true
	entry.Allowed = false
	return entry
}

// GetTrustEntry returns the canonical V2 trust-entry view for a peer.
func (s *Store) GetTrustEntry(
	ctx context.Context,
	peerDID string,
	policy TrustEntryPolicy,
) (*TrustEntry, error) {
	details, err := s.GetDetails(ctx, peerDID)
	if err != nil {
		return nil, err
	}

	entry := EvaluateTrustEntry(peerDID, details, policy)
	return &entry, nil
}

// RecordAdjudicatedFailure records a durable negative outcome after review or
// adjudication. This remains separate from temporary operational incidents.
func (s *Store) RecordAdjudicatedFailure(ctx context.Context, peerDID string) error {
	return s.upsert(ctx, peerDID, func(successes, failures, timeouts int) (int, int, int) {
		return successes, failures + AdjudicatedFailurePenaltyUnits, timeouts
	})
}

// RecordOperationalIncident records a temporary operational safety incident.
// The persistent contract still tracks the event count, but V2 read helpers do
// not treat it as durable negative reputation damage.
func (s *Store) RecordOperationalIncident(ctx context.Context, peerDID string) error {
	return s.RecordTimeout(ctx, peerDID)
}

func clampScore(score float64) float64 {
	switch {
	case score < 0:
		return 0
	case score > 1:
		return 1
	default:
		return score
	}
}
