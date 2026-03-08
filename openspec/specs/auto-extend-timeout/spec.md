# auto-extend-timeout Specification

## Purpose
Configurable automatic deadline extension for agent requests that detects activity (text chunks, tool calls) and extends the timeout up to a maximum cap.
## Requirements
### Requirement: Auto-extend timeout configuration
The system SHALL support `AutoExtendTimeout` (bool) and `MaxRequestTimeout` (duration) fields in `AgentConfig`. When `AutoExtendTimeout` is false (default), behavior SHALL be unchanged.

#### Scenario: Default behavior unchanged
- **WHEN** `AutoExtendTimeout` is not set or false
- **THEN** `runAgent()` SHALL use a fixed `context.WithTimeout` as before

#### Scenario: Auto-extend enabled
- **WHEN** `AutoExtendTimeout` is true
- **THEN** `runAgent()` SHALL use `ExtendableDeadline` instead of fixed timeout

#### Scenario: MaxRequestTimeout defaults to 3x base
- **WHEN** `AutoExtendTimeout` is true and `MaxRequestTimeout` is zero
- **THEN** the maximum timeout SHALL default to 3 times `RequestTimeout`

### Requirement: ExtendableDeadline mechanism
The system SHALL provide an `ExtendableDeadline` that wraps a context with a resettable timer. Each call to `Extend()` resets the deadline by `baseTimeout` from now, but never beyond `maxTimeout` from creation time.

#### Scenario: Expires without extension
- **WHEN** no `Extend()` is called within `baseTimeout`
- **THEN** the context SHALL be canceled after `baseTimeout`

#### Scenario: Extended by activity
- **WHEN** `Extend()` is called before `baseTimeout` expires
- **THEN** the deadline SHALL be reset to `baseTimeout` from the time of the call

#### Scenario: Respects max timeout
- **WHEN** `Extend()` is called repeatedly
- **THEN** the context SHALL be canceled no later than `maxTimeout` from creation time

#### Scenario: Stop cancels immediately
- **WHEN** `Stop()` is called
- **THEN** the context SHALL be canceled immediately

### Requirement: Activity callback in agent runs
The agent `RunAndCollect` and `RunStreaming` methods SHALL accept an optional `WithOnActivity` callback that is invoked on each text chunk or function call event.

#### Scenario: Callback invoked on text event
- **WHEN** the agent produces a text event and `WithOnActivity` is set
- **THEN** the callback SHALL be invoked

#### Scenario: Callback invoked on function call event
- **WHEN** the agent produces a function call event and `WithOnActivity` is set
- **THEN** the callback SHALL be invoked

#### Scenario: No callback when not set
- **WHEN** `WithOnActivity` is not provided
- **THEN** no activity callback SHALL be invoked (no panic or error)

### Requirement: Auto-extend wiring in runAgent
When `AutoExtendTimeout` is enabled, `runAgent()` SHALL wire `WithOnActivity` to call `ExtendableDeadline.Extend()`, so each agent event extends the deadline.

#### Scenario: Agent activity extends deadline
- **WHEN** the agent is actively producing output and `AutoExtendTimeout` is true
- **THEN** the request timeout SHALL be extended on each event up to `MaxRequestTimeout`

