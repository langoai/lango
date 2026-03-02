# Event Bus Spec

## Purpose

Define the synchronous, typed event bus (`internal/eventbus/`) for decoupling callback-based wiring between components.

## Requirements

### Requirement: Event interface
All events SHALL implement an `Event` interface with an `EventName() string` method. The `EventName()` return value is used as the routing key for subscriptions.

#### Scenario: Event routing by name
- **WHEN** an event is published via `Bus.Publish()`
- **THEN** only handlers subscribed to the event's `EventName()` SHALL be invoked

### Requirement: Bus core API
The `Bus` struct SHALL expose `New()`, `Subscribe(eventName, handler)`, and `Publish(event)` methods. `Subscribe` registers a handler for a specific event name. `Publish` dispatches an event to all registered handlers synchronously, in registration order.

#### Scenario: Single handler receives event
- **WHEN** a handler is subscribed to "content.saved" and a `ContentSavedEvent` is published
- **THEN** the handler SHALL be invoked with the event

#### Scenario: Multiple handlers invoked in registration order
- **WHEN** two handlers are subscribed to the same event name
- **THEN** they SHALL be invoked in the order they were subscribed

#### Scenario: Publish with no handlers is no-op
- **WHEN** an event is published and no handlers are registered for that event name
- **THEN** the publish SHALL complete without error (silent no-op)

### Requirement: SubscribeTyped generic helper
The package SHALL provide a `SubscribeTyped[T Event](bus *Bus, handler func(T))` generic function that provides compile-time type safety for event subscriptions.

#### Scenario: Type-safe subscription
- **WHEN** `SubscribeTyped[ContentSavedEvent]` is called with a typed handler
- **THEN** the handler SHALL only be invoked with events of type `ContentSavedEvent`

#### Scenario: Mismatched event type ignored
- **WHEN** a handler subscribed via `SubscribeTyped[ContentSavedEvent]` receives a different event type
- **THEN** the handler SHALL not be invoked

### Requirement: Concurrency safety
`Subscribe` SHALL acquire a write lock. `Publish` SHALL acquire a read lock, copy the handler slice, release the lock, then invoke handlers outside the lock to prevent deadlock from handlers that call `Subscribe`.

#### Scenario: Concurrent publish and subscribe
- **WHEN** multiple goroutines concurrently publish and subscribe
- **THEN** no data race SHALL occur

### Requirement: Event types
The package SHALL define the following event types:

| Event Type              | EventName             | Replaces                                          |
|-------------------------|-----------------------|---------------------------------------------------|
| ContentSavedEvent       | content.saved         | SetEmbedCallback, SetGraphCallback on stores      |
| TriplesExtractedEvent   | triples.extracted     | SetGraphCallback on learning engines/analyzers    |
| TurnCompletedEvent      | turn.completed        | Gateway.OnTurnComplete                            |
| ReputationChangedEvent  | reputation.changed    | reputation.Store.SetOnChangeCallback              |
| MemoryGraphEvent        | memory.graph          | memory.Store.SetGraphHooks                        |

#### Scenario: Each event type has distinct name
- **WHEN** all event types are inspected
- **THEN** each SHALL have a unique `EventName()` return value

### Requirement: Triple type
The package SHALL define a `Triple` struct mirroring `graph.Triple` (Subject, Predicate, Object, Metadata) to avoid importing the graph package, keeping eventbus dependency-free.

#### Scenario: Triple used in TriplesExtractedEvent
- **WHEN** a `TriplesExtractedEvent` is created with triples
- **THEN** the `Triples` field SHALL use `eventbus.Triple` type, not `graph.Triple`

### Requirement: Zero external dependencies
The eventbus package SHALL have zero external dependencies (stdlib only) and SHALL NOT import any other internal package.

#### Scenario: Import validation
- **WHEN** the eventbus package imports are inspected
- **THEN** only standard library packages (e.g., `sync`) SHALL be imported
