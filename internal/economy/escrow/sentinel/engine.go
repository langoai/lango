package sentinel

import (
	"fmt"
	"sync"

	"github.com/langoai/lango/internal/eventbus"
)

// Engine is the Security Sentinel engine that listens to escrow events
// and runs anomaly detectors.
type Engine struct {
	bus       *eventbus.Bus
	config    SentinelConfig
	alerts    []Alert
	mu        sync.RWMutex
	detectors []Detector
	running   bool
	stopCh    chan struct{}
}

// New creates a new Sentinel engine with default detectors.
func New(bus *eventbus.Bus, cfg SentinelConfig) *Engine {
	detectors := []Detector{
		NewRapidCreationDetector(cfg.RapidCreationWindow, cfg.RapidCreationMax),
		NewLargeWithdrawalDetector(cfg.LargeWithdrawalAmount),
		NewRepeatedDisputeDetector(cfg.DisputeWindow, cfg.DisputeMax),
		NewUnusualTimingDetector(cfg.WashTradeWindow),
		NewBalanceDropDetector(),
	}

	return &Engine{
		bus:       bus,
		config:    cfg,
		alerts:    make([]Alert, 0),
		detectors: detectors,
		stopCh:    make(chan struct{}),
	}
}

// Start subscribes to escrow events on the event bus. Idempotent.
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return nil
	}

	e.bus.Subscribe("escrow.created", func(ev eventbus.Event) {
		e.runDetectors(ev)
	})
	e.bus.Subscribe("escrow.released", func(ev eventbus.Event) {
		e.runDetectors(ev)
	})
	e.bus.Subscribe("escrow.milestone", func(ev eventbus.Event) {
		e.runDetectors(ev)
	})

	e.running = true
	return nil
}

// Stop marks the engine as stopped.
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	e.running = false
	close(e.stopCh)
	return nil
}

// runDetectors passes an event through all detectors and collects alerts.
func (e *Engine) runDetectors(event interface{}) {
	for _, d := range e.detectors {
		if alert := d.Analyze(event); alert != nil {
			e.mu.Lock()
			e.alerts = append(e.alerts, *alert)
			e.mu.Unlock()
		}
	}
}

// Alerts returns a copy of all alerts.
func (e *Engine) Alerts() []Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	out := make([]Alert, len(e.alerts))
	copy(out, e.alerts)
	return out
}

// AlertsByLevel returns alerts matching the given severity.
func (e *Engine) AlertsByLevel(severity AlertSeverity) []Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	out := make([]Alert, 0, len(e.alerts))
	for _, a := range e.alerts {
		if a.Severity == severity {
			out = append(out, a)
		}
	}
	return out
}

// ActiveAlerts returns non-acknowledged alerts.
func (e *Engine) ActiveAlerts() []Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	out := make([]Alert, 0, len(e.alerts))
	for _, a := range e.alerts {
		if !a.Acknowledged {
			out = append(out, a)
		}
	}
	return out
}

// Acknowledge marks an alert as acknowledged by ID.
func (e *Engine) Acknowledge(alertID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i := range e.alerts {
		if e.alerts[i].ID == alertID {
			e.alerts[i].Acknowledged = true
			return nil
		}
	}
	return fmt.Errorf("acknowledge alert %q: not found", alertID)
}

// Status returns engine status information.
func (e *Engine) Status() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	detectorNames := make([]string, 0, len(e.detectors))
	for _, d := range e.detectors {
		detectorNames = append(detectorNames, d.Name())
	}

	active := 0
	for _, a := range e.alerts {
		if !a.Acknowledged {
			active++
		}
	}

	return map[string]interface{}{
		"running":      e.running,
		"totalAlerts":  len(e.alerts),
		"activeAlerts": active,
		"detectors":    detectorNames,
	}
}

// Config returns the current sentinel configuration.
func (e *Engine) Config() SentinelConfig {
	return e.config
}
