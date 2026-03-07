package toolchain

import (
	"sync"
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
// It implements both PreToolHook and PostToolHook to measure duration.
// Priority: 50 (runs after security/access checks, observes results).
type EventBusHook struct {
	bus    *eventbus.Bus
	starts sync.Map // key: invocationKey(ctx) -> time.Time
}

// Compile-time interface checks.
var (
	_ PreToolHook  = (*EventBusHook)(nil)
	_ PostToolHook = (*EventBusHook)(nil)
)

// NewEventBusHook creates a new EventBusHook.
func NewEventBusHook(bus *eventbus.Bus) *EventBusHook {
	return &EventBusHook{bus: bus}
}

// Name returns the hook name.
func (h *EventBusHook) Name() string { return "eventbus" }

// Priority returns 50.
func (h *EventBusHook) Priority() int { return 50 }

// Pre records the start time for duration measurement.
func (h *EventBusHook) Pre(ctx HookContext) (PreHookResult, error) {
	h.starts.Store(invocationKey(ctx), time.Now())
	return PreHookResult{Action: Continue}, nil
}

// Post publishes a ToolExecutedEvent to the event bus with measured duration.
func (h *EventBusHook) Post(ctx HookContext, _ interface{}, toolErr error) error {
	var dur time.Duration
	if start, ok := h.starts.LoadAndDelete(invocationKey(ctx)); ok {
		dur = time.Since(start.(time.Time))
	}

	errMsg := ""
	if toolErr != nil {
		errMsg = toolErr.Error()
	}

	h.bus.Publish(ToolExecutedEvent{
		ToolName:   ctx.ToolName,
		AgentName:  ctx.AgentName,
		SessionKey: ctx.SessionKey,
		Duration:   dur,
		Success:    toolErr == nil,
		Error:      errMsg,
	})

	return nil
}

func invocationKey(ctx HookContext) string {
	return ctx.SessionKey + ":" + ctx.ToolName + ":" + ctx.AgentName
}
