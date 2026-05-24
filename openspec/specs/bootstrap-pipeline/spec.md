## Purpose

Capability spec for bootstrap-pipeline. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Pipeline Execute returns Result with phase timing
The `Pipeline.Execute()` method SHALL record the wall-clock duration of each phase and populate `Result.PhaseTiming` with one `PhaseTimingEntry` per completed phase. Each entry SHALL contain the phase name and elapsed duration. A structured log line with `duration_ms` SHALL be emitted after each phase completes.

#### Scenario: All phases succeed
- **WHEN** `Execute()` runs a pipeline with 3 phases that all succeed
- **THEN** `Result.PhaseTiming` contains 3 entries with non-negative durations matching phase names

#### Scenario: A phase fails midway
- **WHEN** the second of 3 phases fails
- **THEN** `Execute()` returns an error and cleanup runs; no timing data is returned (error path)

#### Scenario: Structured log output per phase
- **WHEN** a phase named "openDB" completes in 42ms
- **THEN** a log line is emitted with fields `phase=openDB` and `duration_ms=42`

### Requirement: Bootstrap phase inventory copy stays synchronized

Bootstrap source comments that describe the default phase inventory SHALL stay synchronized with the concrete `DefaultPhases()` sequence.

#### Scenario: Default phase count copy matches implementation

- **WHEN** maintainers read the `DefaultPhases()` source comment
- **THEN** it SHALL state the current 12-phase inventory count

#### Scenario: Run inventory copy lists every default phase

- **WHEN** maintainers read the `Run` source comment
- **THEN** it SHALL list every phase name returned by `DefaultPhases()` in order
- **AND** executable tests SHALL fail if the source comment reintroduces stale phase counts or omits current phase names
