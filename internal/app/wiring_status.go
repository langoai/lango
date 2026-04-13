package app

import (
	"sync"

	"github.com/langoai/lango/internal/types"
)

// StatusCollector aggregates FeatureStatus from wiring functions.
type StatusCollector struct {
	mu       sync.RWMutex
	statuses []types.FeatureStatus
}

// NewStatusCollector creates an empty StatusCollector.
func NewStatusCollector() *StatusCollector {
	return &StatusCollector{}
}

// Add appends a feature status. Nil pointers are ignored.
func (c *StatusCollector) Add(s *types.FeatureStatus) {
	if s == nil {
		return
	}
	c.mu.Lock()
	c.statuses = append(c.statuses, *s)
	c.mu.Unlock()
}

// All returns a copy of all collected statuses.
func (c *StatusCollector) All() []types.FeatureStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]types.FeatureStatus, len(c.statuses))
	copy(out, c.statuses)
	return out
}

// SilentDisabledCount returns the number of features that are disabled
// with a non-empty reason — i.e., features the user likely didn't intend
// to disable but that are off due to missing dependencies.
func (c *StatusCollector) SilentDisabledCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	for _, s := range c.statuses {
		if !s.Enabled && s.Reason != "" {
			count++
		}
	}
	return count
}
