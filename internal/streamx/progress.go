package streamx

import (
	"strings"
	"sync"
)

// ProgressType classifies progress events.
type ProgressType string

const (
	ProgressStarted   ProgressType = "started"
	ProgressUpdate    ProgressType = "update"
	ProgressCompleted ProgressType = "completed"
	ProgressFailed    ProgressType = "failed"
)

// ProgressEvent represents a progress update from any source.
type ProgressEvent struct {
	Source   string         // e.g. "tool:web_search", "agent:operator", "bg:task-123"
	Type     ProgressType
	Message  string         // human-readable progress text
	Progress float64        // 0.0 to 1.0, or -1 if not applicable
	Metadata map[string]any // optional additional data
}

// ProgressBus provides pub/sub for progress events.
type ProgressBus struct {
	mu          sync.RWMutex
	subscribers []*subscriber
}

type subscriber struct {
	filter string
	ch     chan ProgressEvent
	closed bool
}

// NewProgressBus creates a new ProgressBus.
func NewProgressBus() *ProgressBus {
	return &ProgressBus{}
}

// Emit publishes a progress event to all matching subscribers.
// Non-blocking: if a subscriber's buffer is full, the event is dropped.
func (b *ProgressBus) Emit(event ProgressEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, s := range b.subscribers {
		if s.closed {
			continue
		}
		if s.filter != "" && !strings.HasPrefix(event.Source, s.filter) {
			continue
		}
		// Non-blocking send.
		select {
		case s.ch <- event:
		default:
		}
	}
}

// Subscribe returns a channel that receives events matching the filter prefix.
// Call the returned cancel func to unsubscribe and close the channel.
func (b *ProgressBus) Subscribe(filter string) (<-chan ProgressEvent, func()) {
	s := &subscriber{
		filter: filter,
		ch:     make(chan ProgressEvent, 64),
	}

	b.mu.Lock()
	b.subscribers = append(b.subscribers, s)
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if s.closed {
			return
		}
		s.closed = true
		close(s.ch)
		// Remove from list.
		for i, sub := range b.subscribers {
			if sub == s {
				b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
				break
			}
		}
	}
	return s.ch, cancel
}

// SubscribeAll returns a channel receiving all events.
func (b *ProgressBus) SubscribeAll() (<-chan ProgressEvent, func()) {
	return b.Subscribe("")
}
