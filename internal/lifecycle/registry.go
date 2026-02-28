package lifecycle

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

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
	defer r.mu.Unlock()

	var firstErr error
	for i := len(r.started) - 1; i >= 0; i-- {
		if err := r.started[i].Stop(ctx); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("stop %s: %w", r.started[i].Name(), err)
		}
	}
	r.started = nil
	return firstErr
}

// Len returns the number of registered components.
func (r *Registry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.entries)
}
