package streamx

import (
	"context"
	"fmt"
)

// AgentStreamFanIn merges multiple child agent output streams into a single
// tagged stream, emitting progress events for child lifecycle via ProgressBus.
type AgentStreamFanIn struct {
	parent   string                   // parent agent/session ID
	children map[string]Stream[string] // childID -> output stream
	bus      *ProgressBus
}

// NewAgentStreamFanIn creates a fan-in merger for child agent streams.
// If bus is nil, progress emission is skipped.
func NewAgentStreamFanIn(parent string, bus *ProgressBus) *AgentStreamFanIn {
	return &AgentStreamFanIn{
		parent:   parent,
		children: make(map[string]Stream[string]),
		bus:      bus,
	}
}

// AddChild registers a child agent's output stream.
func (f *AgentStreamFanIn) AddChild(childID string, stream Stream[string]) {
	f.children[childID] = stream
}

// MergedStream returns a single stream of tagged events from all children.
// It emits ProgressStarted for each child when merging begins and
// ProgressCompleted/Failed per child as they finish.
func (f *AgentStreamFanIn) MergedStream(ctx context.Context) Stream[Tag[string]] {
	if len(f.children) == 0 {
		return emptyTagStream()
	}

	// Wrap each child stream to emit progress events on completion/error.
	wrapped := make(map[string]Stream[string], len(f.children))
	for id, s := range f.children {
		wrapped[id] = f.wrapChild(id, s)
	}

	// Emit ProgressStarted for each child before merging begins.
	f.emitStarted()

	return Merge[string](ctx, wrapped)
}

// wrapChild returns a stream that delegates to the original, then emits a
// ProgressCompleted or ProgressFailed event when the child stream ends.
func (f *AgentStreamFanIn) wrapChild(childID string, s Stream[string]) Stream[string] {
	return func(yield func(string, error) bool) {
		var childErr error
		for v, err := range s {
			if err != nil {
				childErr = err
				// Forward the error to Merge for propagation.
				yield("", err)
				break
			}
			if !yield(v, nil) {
				// Consumer stopped; treat as normal completion.
				f.emitCompleted(childID)
				return
			}
		}

		if childErr != nil {
			f.emitFailed(childID, childErr)
		} else {
			f.emitCompleted(childID)
		}
	}
}

// emptyTagStream returns a stream that yields nothing.
func emptyTagStream() Stream[Tag[string]] {
	return func(yield func(Tag[string], error) bool) {}
}

func (f *AgentStreamFanIn) emitStarted() {
	if f.bus == nil {
		return
	}
	for id := range f.children {
		f.bus.Emit(ProgressEvent{
			Source:  fmt.Sprintf("agent:%s:child:%s", f.parent, id),
			Type:    ProgressStarted,
			Message: fmt.Sprintf("child agent %s started", id),
		})
	}
}

func (f *AgentStreamFanIn) emitCompleted(childID string) {
	if f.bus == nil {
		return
	}
	f.bus.Emit(ProgressEvent{
		Source:   fmt.Sprintf("agent:%s:child:%s", f.parent, childID),
		Type:     ProgressCompleted,
		Message:  fmt.Sprintf("child agent %s completed", childID),
		Progress: 1.0,
	})
}

func (f *AgentStreamFanIn) emitFailed(childID string, err error) {
	if f.bus == nil {
		return
	}
	f.bus.Emit(ProgressEvent{
		Source:  fmt.Sprintf("agent:%s:child:%s", f.parent, childID),
		Type:    ProgressFailed,
		Message: fmt.Sprintf("child agent %s failed: %v", childID, err),
	})
}
