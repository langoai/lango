## MODIFIED Requirements

### Requirement: Bounded detached trace writes
Trace persistence SHALL use detached contexts with their own timeout so trace writes survive parent cancellation briefly but never block indefinitely. Each create, append, and finish operation SHALL use a fresh detached timeout instead of reusing a single run-scoped timeout context.

#### Scenario: Long turn still records terminal trace state
- **WHEN** a turn runs longer than the configured trace-write timeout
- **THEN** later append and finish operations SHALL still receive a fresh detached timeout context
- **AND** trace persistence SHALL continue attempting to record the terminal outcome independently of earlier trace writes

#### Scenario: Parent cancellation does not lose trace immediately
- **WHEN** the parent request context is cancelled after a failure
- **THEN** the trace writer SHALL continue using a detached context long enough to attempt persistence
- **AND** each detached context SHALL time out independently after the configured trace-write timeout

## ADDED Requirements

### Requirement: Recovery attempts are recorded in turn traces
Structured recovery attempts SHALL be recorded as trace events with enough metadata to identify reroute-vs-retry behavior during diagnosis.

#### Scenario: Specialist reroute recovery is traced
- **WHEN** structured orchestration retries a failed specialist turn with a reroute hint
- **THEN** the trace SHALL append a `recovery_attempt` event
- **AND** the event payload SHALL include the recovery action and failed specialist name

#### Scenario: Generic retry recovery is traced
- **WHEN** structured orchestration retries a turn without a failed specialist identity
- **THEN** the trace SHALL append a `recovery_attempt` event
- **AND** the event payload SHALL distinguish the generic retry from reroute recovery
