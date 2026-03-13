package sentinel

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/eventbus"
)

// Compile-time interface checks.
var (
	_ Detector = (*RapidCreationDetector)(nil)
	_ Detector = (*LargeWithdrawalDetector)(nil)
	_ Detector = (*RepeatedDisputeDetector)(nil)
	_ Detector = (*UnusualTimingDetector)(nil)
	_ Detector = (*BalanceDropDetector)(nil)
)

// windowCounter tracks timestamped events per key within a sliding window.
type windowCounter struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	history map[string][]time.Time
}

// newWindowCounter creates a new window counter.
func newWindowCounter(window time.Duration, max int) windowCounter {
	return windowCounter{
		window:  window,
		max:     max,
		history: make(map[string][]time.Time),
	}
}

// record adds a timestamp for key, prunes entries outside the window, and returns the current count.
func (wc *windowCounter) record(key string) int {
	now := time.Now()
	cutoff := now.Add(-wc.window)

	pruned := make([]time.Time, 0, len(wc.history[key]))
	for _, t := range wc.history[key] {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}
	pruned = append(pruned, now)
	wc.history[key] = pruned

	return len(pruned)
}

// exceeded returns true if the current count for key exceeds max.
func (wc *windowCounter) exceeded(key string) bool {
	return len(wc.history[key]) > wc.max
}

// RapidCreationDetector tracks creation timestamps per peer.
// If more than Max deals from the same peer arrive in Window, it alerts.
type RapidCreationDetector struct {
	windowCounter
}

// NewRapidCreationDetector creates a detector for rapid escrow creation.
func NewRapidCreationDetector(window time.Duration, max int) *RapidCreationDetector {
	return &RapidCreationDetector{windowCounter: newWindowCounter(window, max)}
}

func (d *RapidCreationDetector) Name() string { return "rapid_creation" }

func (d *RapidCreationDetector) Analyze(event interface{}) *Alert {
	ev, ok := event.(eventbus.EscrowCreatedEvent)
	if !ok {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	count := d.record(ev.PayerDID)
	if count > d.max {
		return &Alert{
			ID:        uuid.New().String(),
			Severity:  SeverityHigh,
			Type:      "rapid_creation",
			Message:   fmt.Sprintf("peer %s created %d escrows in %s", ev.PayerDID, count, d.window),
			DealID:    ev.EscrowID,
			PeerDID:   ev.PayerDID,
			Timestamp: time.Now(),
			Metadata:  AlertMetadata{Count: count, Window: d.window.String()},
		}
	}
	return nil
}

// LargeWithdrawalDetector checks release events against a threshold.
type LargeWithdrawalDetector struct {
	mu        sync.Mutex
	threshold *big.Int
}

// NewLargeWithdrawalDetector creates a detector for large withdrawal amounts.
func NewLargeWithdrawalDetector(threshold string) *LargeWithdrawalDetector {
	t := new(big.Int)
	t.SetString(threshold, 10)
	return &LargeWithdrawalDetector{threshold: t}
}

func (d *LargeWithdrawalDetector) Name() string { return "large_withdrawal" }

func (d *LargeWithdrawalDetector) Analyze(event interface{}) *Alert {
	ev, ok := event.(eventbus.EscrowReleasedEvent)
	if !ok {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if ev.Amount == nil || ev.Amount.Cmp(d.threshold) <= 0 {
		return nil
	}

	now := time.Now()
	return &Alert{
		ID:        uuid.New().String(),
		Severity:  SeverityHigh,
		Type:      "large_withdrawal",
		Message:   fmt.Sprintf("large withdrawal of %s from escrow %s", ev.Amount.String(), ev.EscrowID),
		DealID:    ev.EscrowID,
		Timestamp: now,
		Metadata:  AlertMetadata{Amount: ev.Amount.String(), Threshold: d.threshold.String()},
	}
}

// RepeatedDisputeDetector tracks disputes per peer within a window.
type RepeatedDisputeDetector struct {
	windowCounter
}

// NewRepeatedDisputeDetector creates a detector for repeated disputes.
func NewRepeatedDisputeDetector(window time.Duration, max int) *RepeatedDisputeDetector {
	return &RepeatedDisputeDetector{windowCounter: newWindowCounter(window, max)}
}

func (d *RepeatedDisputeDetector) Name() string { return "repeated_dispute" }

func (d *RepeatedDisputeDetector) Analyze(event interface{}) *Alert {
	ev, ok := event.(eventbus.EscrowMilestoneEvent)
	if !ok {
		return nil
	}

	// We use the EscrowID as the peer key here since milestone events
	// don't carry a peer DID. In production the engine would enrich this.
	d.mu.Lock()
	defer d.mu.Unlock()

	peer := ev.EscrowID
	count := d.record(peer)
	if count > d.max {
		return &Alert{
			ID:        uuid.New().String(),
			Severity:  SeverityHigh,
			Type:      "repeated_dispute",
			Message:   fmt.Sprintf("escrow %s triggered %d milestone events in %s", peer, count, d.window),
			DealID:    ev.EscrowID,
			PeerDID:   peer,
			Timestamp: time.Now(),
			Metadata:  AlertMetadata{Count: count, Window: d.window.String()},
		}
	}
	return nil
}

// UnusualTimingDetector detects deals created and released within a short window
// (potential wash trading).
type UnusualTimingDetector struct {
	mu      sync.Mutex
	window  time.Duration
	created map[string]time.Time // escrowID -> creation time
}

// NewUnusualTimingDetector creates a detector for wash-trade-like timing.
func NewUnusualTimingDetector(window time.Duration) *UnusualTimingDetector {
	return &UnusualTimingDetector{
		window:  window,
		created: make(map[string]time.Time),
	}
}

func (d *UnusualTimingDetector) Name() string { return "unusual_timing" }

func (d *UnusualTimingDetector) Analyze(event interface{}) *Alert {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch ev := event.(type) {
	case eventbus.EscrowCreatedEvent:
		d.created[ev.EscrowID] = time.Now()
		return nil

	case eventbus.EscrowReleasedEvent:
		createdAt, ok := d.created[ev.EscrowID]
		if !ok {
			return nil
		}
		delete(d.created, ev.EscrowID)

		elapsed := time.Since(createdAt)
		if elapsed <= d.window {
			now := time.Now()
			return &Alert{
				ID:        uuid.New().String(),
				Severity:  SeverityMedium,
				Type:      "unusual_timing",
				Message:   fmt.Sprintf("escrow %s created and released within %s (possible wash trade)", ev.EscrowID, elapsed.Round(time.Millisecond)),
				DealID:    ev.EscrowID,
				Timestamp: now,
				Metadata:  AlertMetadata{Elapsed: elapsed.String(), Window: d.window.String()},
			}
		}
	}
	return nil
}

// BalanceDropDetector is a placeholder that detects large balance drops.
type BalanceDropDetector struct {
	mu              sync.Mutex
	previousBalance *big.Int
}

// NewBalanceDropDetector creates a detector for significant balance drops.
func NewBalanceDropDetector() *BalanceDropDetector {
	return &BalanceDropDetector{}
}

func (d *BalanceDropDetector) Name() string { return "balance_drop" }

// BalanceChangeEvent can be published externally to feed balance data.
type BalanceChangeEvent struct {
	NewBalance *big.Int
}

func (d *BalanceDropDetector) Analyze(event interface{}) *Alert {
	ev, ok := event.(BalanceChangeEvent)
	if !ok {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.previousBalance == nil || d.previousBalance.Sign() == 0 {
		d.previousBalance = new(big.Int).Set(ev.NewBalance)
		return nil
	}

	// Check if balance dropped by more than 50%.
	half := new(big.Int).Div(d.previousBalance, big.NewInt(2))
	if ev.NewBalance.Cmp(half) < 0 {
		now := time.Now()
		alert := &Alert{
			ID:        uuid.New().String(),
			Severity:  SeverityCritical,
			Type:      "balance_drop",
			Message:   fmt.Sprintf("balance dropped from %s to %s (>50%%)", d.previousBalance.String(), ev.NewBalance.String()),
			Timestamp: now,
			Metadata:  AlertMetadata{PreviousBalance: d.previousBalance.String(), NewBalance: ev.NewBalance.String()},
		}
		d.previousBalance = new(big.Int).Set(ev.NewBalance)
		return alert
	}

	d.previousBalance = new(big.Int).Set(ev.NewBalance)
	return nil
}
