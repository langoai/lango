// Package alerting provides threshold-based operational alerting
// that monitors policy/recovery/budget signals and generates alerts.
package alerting

import (
	"sync"
	"time"

	"github.com/langoai/lango/internal/eventbus"
)

const windowDuration = 5 * time.Minute

// Dispatcher monitors operational signals and publishes AlertEvent
// when configurable thresholds are exceeded within a sliding window.
type Dispatcher struct {
	bus             *eventbus.Bus
	policyThreshold int
	recoveryThresh  int

	mu             sync.Mutex
	policyBlocks   []time.Time
	lastAlertTimes map[string]time.Time
}

// NewDispatcher creates a Dispatcher that publishes alerts to bus
// when the given thresholds are breached within a 5-minute window.
func NewDispatcher(bus *eventbus.Bus, policyBlockRate, recoveryRetries int) *Dispatcher {
	return &Dispatcher{
		bus:             bus,
		policyThreshold: policyBlockRate,
		recoveryThresh:  recoveryRetries,
		lastAlertTimes:  make(map[string]time.Time),
	}
}

// Subscribe registers the dispatcher to receive events from the bus.
func (d *Dispatcher) Subscribe(bus *eventbus.Bus) {
	eventbus.SubscribeTyped[eventbus.PolicyDecisionEvent](bus, d.handlePolicyDecision)
}

func (d *Dispatcher) handlePolicyDecision(evt eventbus.PolicyDecisionEvent) {
	if evt.Verdict != "block" {
		return
	}

	now := time.Now()
	d.mu.Lock()
	defer d.mu.Unlock()

	// Append and prune entries outside the window.
	d.policyBlocks = append(d.policyBlocks, now)
	d.policyBlocks = pruneWindow(d.policyBlocks, now)

	if len(d.policyBlocks) > d.policyThreshold {
		d.maybePublish("policy_block_rate", "warning",
			"policy block rate exceeded threshold",
			map[string]interface{}{
				"count":     len(d.policyBlocks),
				"threshold": d.policyThreshold,
				"window":    windowDuration.String(),
			},
			evt.SessionKey,
			now,
		)
	}
}

// maybePublish publishes an AlertEvent if deduplication allows (one per type per window).
func (d *Dispatcher) maybePublish(alertType, severity, message string, details map[string]interface{}, sessionKey string, now time.Time) {
	if last, ok := d.lastAlertTimes[alertType]; ok && now.Sub(last) < windowDuration {
		return // deduplicated
	}

	d.lastAlertTimes[alertType] = now
	d.bus.Publish(eventbus.AlertEvent{
		Type:       alertType,
		Severity:   severity,
		Message:    message,
		Details:    details,
		SessionKey: sessionKey,
		Timestamp:  now,
	})
}

// pruneWindow removes entries older than windowDuration from a sorted time slice.
func pruneWindow(times []time.Time, now time.Time) []time.Time {
	cutoff := now.Add(-windowDuration)
	i := 0
	for i < len(times) && times[i].Before(cutoff) {
		i++
	}
	if i == 0 {
		return times
	}
	// Shift remaining entries to avoid growing the slice indefinitely.
	copy(times, times[i:])
	return times[:len(times)-i]
}
