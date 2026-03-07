package health

import (
	"context"
	"sync"
	"time"
)

// Registry manages health checkers and performs aggregate checks.
type Registry struct {
	mu       sync.RWMutex
	checkers []Checker
}

// NewRegistry creates a new health Registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a health checker.
func (r *Registry) Register(checker Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers = append(r.checkers, checker)
}

// CheckAll runs all registered checkers and returns an aggregated result.
// The overall status is the worst status among all components.
func (r *Registry) CheckAll(ctx context.Context) SystemHealth {
	r.mu.RLock()
	checkers := make([]Checker, len(r.checkers))
	copy(checkers, r.checkers)
	r.mu.RUnlock()

	components := make([]ComponentHealth, len(checkers))
	for i, c := range checkers {
		components[i] = c.Check(ctx)
	}

	worst := StatusHealthy
	for _, c := range components {
		if statusSeverity(c.Status) > statusSeverity(worst) {
			worst = c.Status
		}
	}

	return SystemHealth{
		Status:     worst,
		Components: components,
		CheckedAt:  time.Now(),
	}
}

// Status returns the current aggregate health status without detailed checks.
func (r *Registry) Status(ctx context.Context) Status {
	return r.CheckAll(ctx).Status
}

func statusSeverity(s Status) int {
	switch s {
	case StatusHealthy:
		return 0
	case StatusDegraded:
		return 1
	case StatusUnhealthy:
		return 2
	default:
		return 3
	}
}
