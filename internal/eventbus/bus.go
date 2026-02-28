// Package eventbus provides a synchronous, typed event bus for decoupling
// components that currently rely on scattered SetXxxCallback() wiring.
//
// Events are dispatched synchronously in registration order. The bus is
// safe for concurrent use; Subscribe takes a write lock, Publish takes a
// read lock.
package eventbus

import "sync"

// Event is implemented by all event types.
type Event interface {
	EventName() string
}

// HandlerFunc processes an event.
type HandlerFunc func(event Event)

// Bus is a synchronous typed event bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]HandlerFunc
}

// New creates a new event bus.
func New() *Bus {
	return &Bus{
		handlers: make(map[string][]HandlerFunc),
	}
}

// Subscribe registers a handler for a specific event name.
func (b *Bus) Subscribe(eventName string, handler HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Publish sends an event to all registered handlers synchronously.
// If no handlers are registered for the event, it is silently ignored.
func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	// Copy the handler slice under the read lock so that a handler calling
	// Subscribe does not deadlock or observe a partially-mutated slice.
	hs := make([]HandlerFunc, len(b.handlers[event.EventName()]))
	copy(hs, b.handlers[event.EventName()])
	b.mu.RUnlock()

	for _, h := range hs {
		h(event)
	}
}

// SubscribeTyped is a generic helper that provides type-safe subscription.
// It registers a handler that automatically type-asserts the event before
// calling the typed handler function.
func SubscribeTyped[T Event](bus *Bus, handler func(T)) {
	var zero T
	bus.Subscribe(zero.EventName(), func(event Event) {
		if typed, ok := event.(T); ok {
			handler(typed)
		}
	})
}
