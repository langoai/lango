package lifecycle

import (
	"context"
	"sync"
)

// Component represents a startable/stoppable application component.
type Component interface {
	Name() string
	Start(ctx context.Context, wg *sync.WaitGroup) error
	Stop(ctx context.Context) error
}

// Priority controls component startup order (lower = earlier).
type Priority int

const (
	PriorityInfra      Priority = 100
	PriorityCore       Priority = 200
	PriorityBuffer     Priority = 300
	PriorityNetwork    Priority = 400
	PriorityAutomation Priority = 500
)

// ComponentEntry pairs a component with its startup priority.
type ComponentEntry struct {
	Component Component
	Priority  Priority
}
