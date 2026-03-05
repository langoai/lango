package toolchain

import (
	"time"

	"github.com/langoai/lango/internal/eventbus"
)

// ToolExecutedEvent is published when a tool finishes execution.
type ToolExecutedEvent struct {
	ToolName   string
	AgentName  string
	SessionKey string
	Duration   time.Duration
	Success    bool
	Error      string
}

// EventName implements eventbus.Event.
func (e ToolExecutedEvent) EventName() string { return "tool.executed" }

// Compile-time interface check.
var _ eventbus.Event = ToolExecutedEvent{}

// EventBusHook publishes tool execution events to the event bus.
// Priority: 50 (runs after security/access checks, observes results).
type EventBusHook struct {
	bus *eventbus.Bus
}

// Compile-time interface check.
var _ PostToolHook = (*EventBusHook)(nil)

// NewEventBusHook creates a new EventBusHook.
func NewEventBusHook(bus *eventbus.Bus) *EventBusHook {
	return &EventBusHook{bus: bus}
}

// Name returns the hook name.
func (h *EventBusHook) Name() string { return "eventbus" }

// Priority returns 50.
func (h *EventBusHook) Priority() int { return 50 }

// Post publishes a ToolExecutedEvent to the event bus.
func (h *EventBusHook) Post(ctx HookContext, _ interface{}, toolErr error) error {
	errMsg := ""
	if toolErr != nil {
		errMsg = toolErr.Error()
	}

	h.bus.Publish(ToolExecutedEvent{
		ToolName:   ctx.ToolName,
		AgentName:  ctx.AgentName,
		SessionKey: ctx.SessionKey,
		Success:    toolErr == nil,
		Error:      errMsg,
	})

	return nil
}
