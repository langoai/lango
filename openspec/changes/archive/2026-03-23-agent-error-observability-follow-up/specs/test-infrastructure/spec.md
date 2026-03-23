## ADDED Requirements

### Requirement: Replay coverage for pre-event failures
The test suite SHALL include a replayable failure fixture for non-success turns that previously produced zero trace events.

#### Scenario: Pre-event failure fixture
- **WHEN** the replay fixture triggers a failure before the first normal runtime event
- **THEN** the resulting trace SHALL still contain a `terminal_error` event

### Requirement: Table-driven cause-class verification
The test suite SHALL verify the initial failure-cause taxonomy via table-driven tests.

#### Scenario: Cause-class table test
- **WHEN** the classification tests run
- **THEN** approval, tool lookup, tool validation, provider, timeout, turn limit, repeated-call, and empty-after-tool-use cases SHALL map to their expected `CauseClass` values
