## MODIFIED Requirements

### Requirement: Build returns BuildResult with module timing
The `Builder.Build()` method SHALL record the wall-clock duration of each module's `Init()` call and populate `BuildResult.ModuleTiming` with one `ModuleTimingEntry` per initialized module. Each entry SHALL contain the module name and elapsed duration. A structured log line with `duration_ms` SHALL be emitted after each module completes initialization.

#### Scenario: All modules initialize successfully
- **WHEN** `Build()` initializes 3 modules that all succeed
- **THEN** `BuildResult.ModuleTiming` contains 3 entries with non-negative durations matching module names

#### Scenario: A module fails
- **WHEN** the second module's `Init()` returns an error
- **THEN** `Build()` returns an error; no timing data is returned

#### Scenario: Structured log output per module
- **WHEN** a module named "knowledge" initializes in 85ms
- **THEN** a log line is emitted with fields `module=knowledge` and `duration_ms=85`
