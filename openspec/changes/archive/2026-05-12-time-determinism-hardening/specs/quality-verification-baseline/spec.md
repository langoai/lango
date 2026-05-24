## ADDED Requirements

### Requirement: Clock-sensitive regression tests use deterministic time sources
The repository SHALL keep TTL- and age-threshold regression tests deterministic by using execution-relative timestamps or an explicitly injected clock whenever the asserted behavior depends on elapsed time.

#### Scenario: Proposal lifecycle regression uses execution-relative timestamps
- **WHEN** the proposal module regression constructs learning suggestion events for TTL-sensitive assertions
- **THEN** it derives those timestamps from the current execution window instead of a historical fixed calendar date

#### Scenario: Mission Control agenda regression pins the projector clock
- **WHEN** a Mission Control regression asserts cron failure freshness or inquiry follow-up age thresholds
- **THEN** it injects the projector clock used by the threshold calculation before asserting loop ordering or visibility

### Requirement: Parallel read-only executor reports positive duration for completed invocations
The parallel read-only executor SHALL record a positive duration for every completed eligible invocation, including handlers that return an error.

#### Scenario: Failing handler still records duration
- **WHEN** an eligible parallel read-only tool handler returns an error after execution begins
- **THEN** the corresponding `ToolResult` includes that error and a duration greater than zero
