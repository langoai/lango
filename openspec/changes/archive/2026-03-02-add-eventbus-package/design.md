# Design: Event Bus Package

## Architecture

```
internal/eventbus/
├── bus.go          # Bus struct, New, Subscribe, Publish, SubscribeTyped
├── events.go       # Concrete event types and Triple mirror type
└── bus_test.go     # Comprehensive test suite
```

## Key Decisions

### Synchronous Dispatch

All current callbacks are synchronous. The event bus preserves this behavior
to ensure drop-in replacement compatibility. Async dispatch can be added later
as a separate `PublishAsync` method if needed.

### Handler Slice Copy on Publish

`Publish` copies the handler slice under the read lock before invoking handlers.
This prevents deadlock when a handler calls `Subscribe` during dispatch and
avoids observing a partially-mutated slice.

### Generic SubscribeTyped

Uses Go generics to provide type-safe subscriptions:

```go
SubscribeTyped(bus, func(e ContentSavedEvent) {
    // e is already typed, no assertion needed
})
```

Internally creates a `HandlerFunc` that type-asserts before calling the typed
handler. The event name is derived from the zero value of the type parameter.

### Mirror Types

`eventbus.Triple` mirrors `graph.Triple` to keep the eventbus package
completely dependency-free. Conversion between the two types will happen at
the boundary (in wiring code) during the migration phase.

## Thread Safety

| Operation  | Lock      | Notes                                      |
|------------|-----------|--------------------------------------------|
| Subscribe  | Write     | Appends to handler slice                   |
| Publish    | Read      | Copies slice under lock, invokes outside   |

This allows concurrent publishes (common path) while serializing subscriptions
(rare path, typically at startup).

## Test Coverage

- Single handler receives published event
- Multiple handlers receive in registration order
- No panic on publish with no handlers
- SubscribeTyped type-safe handling
- Different event types route to different handlers
- Concurrent publish/subscribe (race detector)
- All event types have distinct names
- Round-trip tests for each concrete event type
