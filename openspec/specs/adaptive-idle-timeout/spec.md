# adaptive-idle-timeout Specification

## Purpose
TBD - created by archiving change adaptive-idle-timeout. Update Purpose after archive.
## Requirements
### Requirement: Shared ExtendableDeadline package
The system SHALL provide an `internal/deadline` package containing `ExtendableDeadline` with `New()`, `Extend()`, `Stop()`, and `Reason()` methods. The `internal/app` package SHALL re-export via type alias for backward compatibility.

#### Scenario: Idle timeout fires on inactivity
- **WHEN** an ExtendableDeadline is created with idleTimeout=2m and no Extend() calls occur
- **THEN** the context SHALL be cancelled after 2 minutes with Reason() returning "idle"

#### Scenario: Activity extends the idle timer
- **WHEN** Extend() is called before the idle timeout expires
- **THEN** the idle timer SHALL reset to idleTimeout from the current time

#### Scenario: Hard ceiling is enforced
- **WHEN** Extend() is called repeatedly and the total elapsed time reaches maxTimeout
- **THEN** the context SHALL be cancelled with Reason() returning "max_timeout"

#### Scenario: Stop cancels immediately
- **WHEN** Stop() is called
- **THEN** the context SHALL be cancelled immediately with Reason() returning "cancelled"

### Requirement: IdleTimeout config field
The `AgentConfig` SHALL include an `IdleTimeout` field of type `time.Duration`. When positive, idle timeout mode is active. When negative (-1), idle timeout is explicitly disabled. When zero (default), behavior depends on other config fields.

The `deadline` package SHALL provide a `ResolveTimeouts(cfg TimeoutConfig) (idleTimeout, hardCeiling time.Duration)` function that encapsulates all timeout resolution logic. Both channel handlers (`runAgent`) and the gateway (`initGateway`) SHALL use this function as the single source of truth for timeout computation.

#### Scenario: IdleTimeout set to 2m
- **WHEN** config has `idleTimeout: 2m` and `requestTimeout: 30m`
- **THEN** `deadline.ResolveTimeouts()` SHALL return idle=2m, ceiling=30m

#### Scenario: IdleTimeout not set
- **WHEN** config has only `requestTimeout: 5m`
- **THEN** `deadline.ResolveTimeouts()` SHALL return idle=0, ceiling=5m (fixed timeout, backward compatible)

#### Scenario: IdleTimeout set to -1
- **WHEN** config has `idleTimeout: -1`
- **THEN** `deadline.ResolveTimeouts()` SHALL return idle=0 (disabled), using fixed RequestTimeout

#### Scenario: Legacy AutoExtendTimeout
- **WHEN** config has `autoExtendTimeout: true` with `requestTimeout: 5m` and `maxRequestTimeout: 15m`
- **THEN** `deadline.ResolveTimeouts()` SHALL return idle=5m, ceiling=15m

#### Scenario: Gateway uses same resolution as channels
- **WHEN** `initGateway()` constructs gateway config
- **THEN** it SHALL call `deadline.ResolveTimeouts()` with the same `TimeoutConfig` fields used by `runAgent()`

### Requirement: Idle timeout in channel handlers
The `runAgent()` method SHALL use `resolveTimeouts()` to determine timeout behavior. When idle timeout is active, it SHALL create an `ExtendableDeadline` and wire `WithOnActivity` to call `Extend()`.

#### Scenario: Active agent extends deadline
- **WHEN** the agent produces streaming chunks or tool calls and idle timeout is active
- **THEN** the idle timer SHALL be extended on each activity event

#### Scenario: Stalled agent times out
- **WHEN** the agent produces no activity for the idle timeout duration
- **THEN** the request SHALL be cancelled with ErrIdleTimeout (E006)

### Requirement: Idle timeout in gateway
The gateway `handleChatMessage()` SHALL support idle timeout via `Config.IdleTimeout` and `Config.MaxTimeout` fields. When `IdleTimeout > 0`, it SHALL use `ExtendableDeadline` and pass `WithOnActivity` to `RunStreaming`. Error types in gateway timeout handling SHALL use `string(deadline.ReasonXxx)` constants instead of raw string literals.

#### Scenario: Gateway idle timeout fires
- **WHEN** gateway idle timeout is active and no agent activity occurs
- **THEN** the gateway SHALL cancel the request and broadcast an "agent.error" event with type `string(deadline.ReasonIdle)` (value: "idle")

#### Scenario: Gateway max timeout fires
- **WHEN** gateway max timeout is reached
- **THEN** the gateway SHALL broadcast an "agent.error" event with type `string(deadline.ReasonMaxTimeout)` (value: "max_timeout")

### Requirement: Session timeout annotation
The `session.Store` interface SHALL include `AnnotateTimeout(key string, partial string) error`. On timeout, callers SHALL invoke this to append a synthetic assistant message marking the interrupted turn.

#### Scenario: Timeout with no partial response
- **WHEN** a timeout occurs and no partial text was accumulated
- **THEN** AnnotateTimeout SHALL append an assistant message with "[This response was interrupted due to a timeout]"

#### Scenario: Timeout with partial response
- **WHEN** a timeout occurs and partial text was accumulated
- **THEN** AnnotateTimeout SHALL append an assistant message containing the partial text followed by the timeout marker

#### Scenario: Next turn after timeout
- **WHEN** the user sends a new message after a timeout-annotated turn
- **THEN** the session history SHALL contain a complete user→assistant pair, preventing error leakage

### Requirement: ErrIdleTimeout error code
The ADK error system SHALL include `ErrIdleTimeout` (E006) for idle-specific timeouts, distinct from the general `ErrTimeout` (E001) used for hard ceiling timeouts.

#### Scenario: Idle timeout error message
- **WHEN** an idle timeout occurs
- **THEN** the user-facing message SHALL indicate the inactivity duration

