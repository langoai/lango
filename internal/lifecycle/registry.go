package lifecycle

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/langoai/lango/internal/logging"
)

var logger = logging.SubsystemSugar("lifecycle")

// Registry manages component lifecycle with ordered startup and reverse shutdown.
type Registry struct {
	mu      sync.Mutex
	entries []ComponentEntry
	started []Component
}

// NewRegistry creates an empty component registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a component at the given priority.
func (r *Registry) Register(c Component, p Priority) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, ComponentEntry{Component: c, Priority: p})
}

// StartAll starts all registered components in priority order (ascending).
// If a component fails to start, already-started components are stopped in
// reverse order (rollback).
func (r *Registry) StartAll(ctx context.Context, wg *sync.WaitGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sorted := make([]ComponentEntry, len(r.entries))
	copy(sorted, r.entries)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	r.started = r.started[:0]

	for _, entry := range sorted {
		if err := entry.Component.Start(ctx, wg); err != nil {
			for i := len(r.started) - 1; i >= 0; i-- {
				_ = r.started[i].Stop(ctx)
			}
			r.started = nil
			return fmt.Errorf("start %s: %w", entry.Component.Name(), err)
		}
		r.started = append(r.started, entry.Component)
	}

	return nil
}

// StopAll stops all started components in reverse startup order.
func (r *Registry) StopAll(ctx context.Context) error {
	r.mu.Lock()
	started := make([]Component, len(r.started))
	copy(started, r.started)
	r.started = nil
	r.mu.Unlock()

	var firstErr error
	for i := len(started) - 1; i >= 0; i-- {
		component := started[i]
		logger.Infow("stopping component", "component", component.Name())

		done := make(chan error, 1)
		go func(c Component) {
			done <- c.Stop(ctx)
		}(component)

		select {
		case err := <-done:
			if err != nil {
				logger.Warnw("component stop error", "component", component.Name(), "error", err)
				if firstErr == nil {
					firstErr = fmt.Errorf("stop %s: %w", component.Name(), err)
				}
				continue
			}
			logger.Infow("stopped component", "component", component.Name())
		case <-ctx.Done():
			logger.Warnw("component stop timed out", "component", component.Name(), "error", ctx.Err())
			if firstErr == nil {
				firstErr = fmt.Errorf("stop %s: %w", component.Name(), ctx.Err())
			}
		}
	}
	return firstErr
}

// Len returns the number of registered components.
func (r *Registry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.entries)
}

// Names returns the names of all registered components in registration order.
func (r *Registry) Names() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	names := make([]string, len(r.entries))
	for i, e := range r.entries {
		names[i] = e.Component.Name()
	}
	return names
}
