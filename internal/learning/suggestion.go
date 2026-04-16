package learning

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/langoai/lango/internal/eventbus"
)

// SuggestionEmitter decides whether a learning candidate should surface to
// the user as an approval-gated suggestion, and publishes a
// LearningSuggestionEvent when it does. Rate-limiting is per-session; dedup
// is global by pattern hash within a sliding window.
type SuggestionEmitter struct {
	bus            *eventbus.Bus
	threshold      float64
	rateLimit      int
	driftThreshold int
	dedupWindow    time.Duration
	now            func() time.Time

	mu             sync.Mutex
	turnCounters   map[string]int       // sessionKey -> turns since last emit
	recentHashes   map[string]time.Time // pattern hash -> last emit time
	dismissed      map[string]time.Time // pattern hash -> dismissal time (also serves as negative dedup)
	driftCounters  map[string]driftEntry // "toolName:errorClass" -> count + first seen
}

type driftEntry struct {
	count     int
	firstSeen time.Time
}

// SuggestionCandidate is the minimum shape an emitter needs to decide and
// serialize a suggestion.
type SuggestionCandidate struct {
	SessionKey   string
	Pattern      string
	ProposedRule string
	Confidence   float64
	Rationale    string
}

const defaultDriftThreshold = 5

// NewSuggestionEmitter constructs an emitter. bus may be nil — callers can
// use this emitter in tests without wiring the event bus.
func NewSuggestionEmitter(bus *eventbus.Bus, threshold float64, rateLimit int, dedupWindow time.Duration) *SuggestionEmitter {
	return &SuggestionEmitter{
		bus:            bus,
		threshold:      threshold,
		rateLimit:      rateLimit,
		driftThreshold: defaultDriftThreshold,
		dedupWindow:    dedupWindow,
		now:            time.Now,
		turnCounters:   make(map[string]int),
		recentHashes:   make(map[string]time.Time),
		dismissed:      make(map[string]time.Time),
		driftCounters:  make(map[string]driftEntry),
	}
}

// TickTurn increments the per-session turn counter. Call once at
// TurnCompletedEvent. The rate-limit decision is made against this count.
func (e *SuggestionEmitter) TickTurn(sessionKey string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.turnCounters[sessionKey]++
}

// MaybeEmit evaluates a candidate and publishes if it passes threshold,
// rate-limit, and dedup checks. Returns true if an event was published.
func (e *SuggestionEmitter) MaybeEmit(_ context.Context, c SuggestionCandidate) bool {
	if c.Confidence < e.threshold {
		return false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.now()

	hash := patternHash(c.Pattern)

	// Dedup: pattern hash seen recently (including dismissal)?
	if t, ok := e.recentHashes[hash]; ok && now.Sub(t) < e.dedupWindow {
		return false
	}
	if t, ok := e.dismissed[hash]; ok && now.Sub(t) < e.dedupWindow {
		return false
	}

	// Rate limit per session.
	if turns, ok := e.turnCounters[c.SessionKey]; ok {
		if turns < e.rateLimit {
			return false
		}
	}

	// Publish and reset counter.
	e.recentHashes[hash] = now
	e.turnCounters[c.SessionKey] = 0

	if e.bus != nil {
		e.bus.Publish(eventbus.LearningSuggestionEvent{
			SessionKey:   c.SessionKey,
			SuggestionID: hash,
			Pattern:      c.Pattern,
			ProposedRule: c.ProposedRule,
			Confidence:   c.Confidence,
			Rationale:    c.Rationale,
			Timestamp:    now,
		})
	}
	return true
}

// EmitSpecDrift tracks recurring error patterns per tool and publishes a
// SpecDriftDetectedEvent when the occurrence threshold is crossed. Returns
// true if an event was published.
func (e *SuggestionEmitter) EmitSpecDrift(_ context.Context, toolName, errorClass, sampleErr string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.now()

	key := fmt.Sprintf("%s:%s", toolName, errorClass)
	hash := fmt.Sprintf("drift-%x", sha256.Sum256([]byte(key)))[:20]

	if t, ok := e.recentHashes[hash]; ok && now.Sub(t) < e.dedupWindow {
		return false
	}

	entry := e.driftCounters[key]
	if entry.firstSeen.IsZero() {
		entry.firstSeen = now
	}
	if now.Sub(entry.firstSeen) >= e.dedupWindow {
		entry = driftEntry{firstSeen: now}
	}
	entry.count++
	e.driftCounters[key] = entry

	if entry.count < e.driftThreshold {
		return false
	}

	occurrences := entry.count
	delete(e.driftCounters, key)
	e.recentHashes[hash] = now

	if e.bus != nil {
		e.bus.Publish(eventbus.SpecDriftDetectedEvent{
			ToolName:    toolName,
			ErrorClass:  errorClass,
			Occurrences: occurrences,
			SampleError: sampleErr,
			Timestamp:   now,
		})
	}
	return true
}

// Dismiss records that a suggestion was denied by the user. Suppresses
// re-emission of the same pattern within dedupWindow.
func (e *SuggestionEmitter) Dismiss(pattern string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.dismissed[patternHash(pattern)] = e.now()
}

// Prune removes entries older than dedupWindow. Callers can invoke this
// periodically to keep map sizes bounded in long-running processes.
func (e *SuggestionEmitter) Prune() {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.now()
	for k, t := range e.recentHashes {
		if now.Sub(t) >= e.dedupWindow {
			delete(e.recentHashes, k)
		}
	}
	for k, entry := range e.driftCounters {
		if now.Sub(entry.firstSeen) >= e.dedupWindow {
			delete(e.driftCounters, k)
		}
	}
	for k, t := range e.dismissed {
		if now.Sub(t) >= e.dedupWindow {
			delete(e.dismissed, k)
		}
	}
}

func patternHash(pattern string) string {
	sum := sha256.Sum256([]byte(pattern))
	return fmt.Sprintf("sugg-%s", hex.EncodeToString(sum[:8]))
}
