## MODIFIED Requirements

### Requirement: ExtendableDeadline mechanism
The system SHALL provide an `ExtendableDeadline` in the `internal/deadline` package (extracted from `internal/app`) that wraps a context with a resettable idle timer. Each call to `Extend()` resets the deadline by `idleTimeout` from now, but never beyond `maxTimeout` from creation time. The type SHALL expose a `Reason()` method returning the cause of expiry: `"idle"`, `"max_timeout"`, or `"cancelled"`.

#### Scenario: Expires without extension
- **WHEN** no `Extend()` is called within `idleTimeout`
- **THEN** the context SHALL be canceled after `idleTimeout` and `Reason()` SHALL return `"idle"`

#### Scenario: Extended by activity
- **WHEN** `Extend()` is called before `idleTimeout` expires
- **THEN** the deadline SHALL be reset to `idleTimeout` from the time of the call

#### Scenario: Respects max timeout
- **WHEN** `Extend()` is called repeatedly
- **THEN** the context SHALL be canceled no later than `maxTimeout` from creation time and `Reason()` SHALL return `"max_timeout"`

#### Scenario: Stop cancels immediately
- **WHEN** `Stop()` is called
- **THEN** the context SHALL be canceled immediately and `Reason()` SHALL return `"cancelled"`

#### Scenario: Backward-compatible alias
- **WHEN** code in `internal/app` references `ExtendableDeadline` or `NewExtendableDeadline`
- **THEN** the type alias and wrapper function SHALL delegate to `internal/deadline` without behavioral changes

### Requirement: Auto-extend wiring in runAgent
When idle timeout is active (via `IdleTimeout > 0` or legacy `AutoExtendTimeout = true`), `runAgent()` SHALL use `resolveTimeouts()` to determine idle and ceiling values, create an `ExtendableDeadline`, and wire `WithOnActivity` to call `Extend()`.

#### Scenario: IdleTimeout config takes precedence
- **WHEN** `IdleTimeout` is set to a positive duration
- **THEN** `resolveTimeouts()` SHALL use it as the idle timeout regardless of `AutoExtendTimeout`

#### Scenario: Legacy AutoExtendTimeout mapping
- **WHEN** `AutoExtendTimeout` is true and `IdleTimeout` is zero
- **THEN** `resolveTimeouts()` SHALL map `RequestTimeout` as idle and `MaxRequestTimeout` as ceiling

#### Scenario: Default fixed timeout preserved
- **WHEN** neither `IdleTimeout` nor `AutoExtendTimeout` is set
- **THEN** `resolveTimeouts()` SHALL return idle=0 with `RequestTimeout` as a fixed ceiling
