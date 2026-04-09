## Purpose

Capability spec for progress-bus. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Progress event types and structure
The system SHALL define a `ProgressEvent` struct with `Source` (string), `Type` (ProgressType), `Message` (string), `Progress` (float64), and `Metadata` (map[string]any). `ProgressType` SHALL be a string type with constants `started`, `update`, `completed`, and `failed`.

#### Scenario: Event with all fields
- **WHEN** a `ProgressEvent` is created with Source="tool:web_search", Type=ProgressStarted, Message="searching", Progress=0.0, Metadata={"query": "test"}
- **THEN** all fields SHALL be accessible and retain their assigned values

#### Scenario: Progress value semantics
- **WHEN** `Progress` is set to -1
- **THEN** it SHALL indicate an indeterminate progress state

### Requirement: Subscribe with prefix filter
`ProgressBus.Subscribe(filter)` SHALL return a receive-only channel (buffered, capacity 64) and a cancel function. Only events whose `Source` starts with the given `filter` prefix SHALL be delivered to the channel.

#### Scenario: Prefix filtering delivers matching events
- **WHEN** a subscriber with filter "tool:" exists and an event with Source="tool:web_search" is emitted
- **THEN** the event SHALL be delivered to the subscriber's channel

#### Scenario: Prefix filtering excludes non-matching events
- **WHEN** a subscriber with filter "tool:" exists and an event with Source="agent:operator" is emitted
- **THEN** the event SHALL NOT be delivered to the subscriber's channel

### Requirement: SubscribeAll receives all events
`ProgressBus.SubscribeAll()` SHALL return a channel that receives all emitted events regardless of source.

#### Scenario: SubscribeAll receives events from all sources
- **WHEN** a SubscribeAll subscriber exists and events with Source="tool:fs_read" and Source="agent:vault" are emitted
- **THEN** both events SHALL be delivered to the subscriber's channel

### Requirement: Cancel closes channel and removes subscriber
The cancel function returned by `Subscribe` SHALL close the subscriber's channel and remove the subscriber from the bus. Double-cancel SHALL be safe (no-op). Emitting after cancel SHALL not panic.

#### Scenario: Cancel closes channel
- **WHEN** the cancel function is called
- **THEN** the subscriber's channel SHALL be closed and subsequent receives SHALL return the zero value with ok=false

#### Scenario: Double cancel is safe
- **WHEN** the cancel function is called twice
- **THEN** the second call SHALL be a no-op without panic

#### Scenario: Emit after cancel is safe
- **WHEN** an event is emitted after a subscriber has been cancelled
- **THEN** the emit SHALL complete without panic

### Requirement: Buffer full drops events (non-blocking)
`ProgressBus.Emit()` SHALL use non-blocking sends. **WHEN** a subscriber's channel buffer (capacity 64) is full, the event SHALL be silently dropped for that subscriber without blocking the emitter.

#### Scenario: Buffer overflow drops events
- **WHEN** more than 64 events are emitted without the subscriber consuming any
- **THEN** only the first 64 events SHALL be buffered and subsequent events SHALL be dropped

### Requirement: Concurrent emit safety
`ProgressBus.Emit()` SHALL be safe for concurrent use from multiple goroutines. Multiple goroutines SHALL be able to emit events simultaneously without data races.

#### Scenario: Concurrent emit from multiple goroutines
- **WHEN** 10 goroutines each emit one event concurrently
- **THEN** all events SHALL be delivered without data races and the subscriber SHALL receive all 10 events (assuming buffer capacity is sufficient)
