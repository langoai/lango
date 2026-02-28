## ADDED Requirements

### Requirement: Event interface
The system SHALL define an Event interface with EventName() string for typed event identification.

#### Scenario: Event returns its name
- **WHEN** a ContentSavedEvent is created
- **THEN** EventName() SHALL return "content.saved"

### Requirement: Subscribe and publish
The Bus SHALL support Subscribe(eventName, handler) and Publish(event), calling all handlers registered for the event's name synchronously in registration order.

#### Scenario: Multiple handlers receive event
- **WHEN** two handlers are subscribed to "turn.completed" and a TurnCompletedEvent is published
- **THEN** both handlers SHALL be called in registration order

#### Scenario: No handlers registered
- **WHEN** an event is published with no subscribers
- **THEN** the event SHALL be silently ignored without error or panic

### Requirement: Type-safe subscription
The system SHALL provide SubscribeTyped[T Event] for generic type-safe event handling without manual type assertions.

#### Scenario: Typed handler receives correct type
- **WHEN** SubscribeTyped[TurnCompletedEvent] is used with a handler
- **THEN** the handler SHALL receive TurnCompletedEvent directly (not Event interface)

### Requirement: Thread safety
The Bus SHALL be safe for concurrent Subscribe and Publish calls.

#### Scenario: Concurrent publish and subscribe
- **WHEN** multiple goroutines publish and subscribe simultaneously
- **THEN** no data races SHALL occur (verified by -race flag)

### Requirement: Content event types
The system SHALL define ContentSavedEvent, TriplesExtractedEvent, TurnCompletedEvent, ReputationChangedEvent, and MemoryGraphEvent as concrete Event implementations.

#### Scenario: Each event has unique name
- **WHEN** all event types are instantiated
- **THEN** each SHALL have a unique EventName() value
