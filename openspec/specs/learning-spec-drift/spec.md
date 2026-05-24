# Learning Spec Drift Detection

## Purpose
Detect when accumulated learning patterns indicate OpenSpec documentation may be stale and signal via EventBus without generating artifacts.

## Requirements

### Requirement: SpecDriftDetectedEvent
The eventbus SHALL define a `SpecDriftDetectedEvent` containing the tool name, error class, occurrence count, a sample error message, and an optional affected spec name. The event SHALL be published through the existing EventBus infrastructure.

#### Scenario: Event published on recurring error pattern
- **WHEN** the same `(toolName, errorClass)` pair recurs at least N times (default 5) within the dedup window
- **THEN** a `SpecDriftDetectedEvent` is published via EventBus
- **AND** the event includes the tool name, error class, occurrence count, and a sample error

#### Scenario: One-off errors do not trigger drift
- **WHEN** a tool error occurs fewer than N times
- **THEN** no `SpecDriftDetectedEvent` is published

#### Scenario: Dedup prevents repeated drift events
- **WHEN** a drift event has already been published for a `(toolName, errorClass)` pair within the dedup window
- **THEN** no duplicate event is published

### Requirement: EmitSpecDrift method on SuggestionEmitter
The `SuggestionEmitter` SHALL expose an `EmitSpecDrift(ctx, toolName, errorClass, sampleErr)` method that tracks error pattern frequency and publishes `SpecDriftDetectedEvent` when the threshold is crossed. The method SHALL reuse the emitter's existing dedup and rate-limit infrastructure.

#### Scenario: Threshold crossed triggers emission
- **WHEN** `EmitSpecDrift` is called with the same `(toolName, errorClass)` for the Nth time
- **THEN** a `SpecDriftDetectedEvent` is published
- **AND** the internal counter for that pair is reset

#### Scenario: Below threshold accumulates silently
- **WHEN** `EmitSpecDrift` is called fewer than N times for a pair
- **THEN** no event is published
- **AND** the counter is incremented

### Requirement: Integration with learning engine
The learning engine's `OnToolResult` path SHALL call `EmitSpecDrift` when a tool result includes an error, passing the tool name, classified error cause, and error message.

#### Scenario: Tool error triggers drift tracking
- **WHEN** `OnToolResult` is called with a non-nil error
- **THEN** `EmitSpecDrift` is called with the tool name and error classification
