# Spec: Event Bus

## Overview

`internal/eventbus/` provides a synchronous, typed event bus for decoupling
callback-based wiring between components.

## Interface

### Event

```go
type Event interface {
    EventName() string
}
```

All events must implement this interface. The `EventName()` return value is
used as the routing key for subscriptions.

### Bus

```go
type Bus struct { ... }

func New() *Bus
func (b *Bus) Subscribe(eventName string, handler HandlerFunc)
func (b *Bus) Publish(event Event)
func SubscribeTyped[T Event](bus *Bus, handler func(T))
```

- `Subscribe` registers a handler for a specific event name.
- `Publish` dispatches an event to all registered handlers synchronously, in
  registration order.
- `SubscribeTyped` is a generic helper that provides compile-time type safety.

### Concurrency

- `Subscribe` acquires a write lock.
- `Publish` acquires a read lock, copies the handler slice, releases the lock,
  then invokes handlers outside the lock to prevent deadlock from handlers
  that call `Subscribe`.

## Event Types

| Event Type              | EventName             | Replaces                                          |
|-------------------------|-----------------------|---------------------------------------------------|
| ContentSavedEvent       | content.saved         | SetEmbedCallback, SetGraphCallback on stores      |
| TriplesExtractedEvent   | triples.extracted     | SetGraphCallback on learning engines/analyzers    |
| TurnCompletedEvent      | turn.completed        | Gateway.OnTurnComplete                            |
| ReputationChangedEvent  | reputation.changed    | reputation.Store.SetOnChangeCallback              |
| MemoryGraphEvent        | memory.graph          | memory.Store.SetGraphHooks                        |

### Triple Type

`eventbus.Triple` mirrors `graph.Triple` to avoid importing the graph package,
keeping eventbus dependency-free.

## Constraints

- Zero external dependencies (stdlib only)
- No import of any other internal package
- Handlers are invoked synchronously in registration order
- No handler is registered for an event: publish is a silent no-op
