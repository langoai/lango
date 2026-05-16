## Purpose

Capability spec for system-feedback. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Non-blocking startup security
The system SHALL prioritize environment-based security credentials (LANGO_PASSPHRASE) over interactive prompts to ensure automated and remote startups are not blocked.

#### Scenario: Passphrase provided via environment
- **WHEN** the `LANGO_PASSPHRASE` environment variable is set
- **THEN** the system SHALL use it and SKIP any interactive prompts, even in a TTY environment.

#### Scenario: Passphrase missing in interactive session
- **WHEN** `LANGO_PASSPHRASE` is NOT set AND the session is interactive (TTY)
- **THEN** the system SHALL prompt the user for the passphrase.

#### Scenario: Passphrase missing in headless environment
- **WHEN** `LANGO_PASSPHRASE` is NOT set AND the session is NOT interactive
- **THEN** the system SHALL terminate with a descriptive error.

### Requirement: System lifecycle visibility
The system SHALL provide granular logging during its startup and initialization phase to inform the user of the status of each core component.

#### Scenario: Component initialization feedback
- **WHEN** a major component (Supervisor, Agent, Gateway, Channel) starts initializing
- **THEN** the system SHALL log its progress and reporting any success or failure immediately.

#### Scenario: Gateway and Bot readiness
- **WHEN** all components are successfully initialized and listening
- **THEN** the system SHALL log a clear "Ready" message including the server address and names of active bot channels.

### Requirement: Observe-only graph admission telemetry is emitted for runtime feedback
The runtime SHALL emit observe-only graph admission telemetry and non-admission baseline signals for the supported runtime graph inputs in this slice.

#### Scenario: Supported inputs emit telemetry on the runtime event bus
- **WHEN** observe-only admission processes a supported runtime graph input
- **THEN** event-bus admission telemetry SHALL preserve producer-source, producer-group, and validator-source identity
- **AND** `content_saved_extractor` telemetry SHALL preserve its synthetic source label and validator-source identity without inventing a producer-group

#### Scenario: Non-admission baseline signals remain distinct
- **WHEN** the runtime emits dropped-unknown, `unmapped-source`, or graph write-failure baseline feedback
- **THEN** those signals SHALL remain distinct from graph-admission telemetry
- **AND** they SHALL preserve the source identities required by their respective contracts

#### Scenario: Validator-unavailable mode emits unvalidated admission observations
- **WHEN** ontology is disabled or ontology initialization fails before observe-only admission can classify a batch through the shared validator closure
- **THEN** the runtime SHALL emit an `unvalidated` graph-admission observation for that batch with `validator_source = "unavailable"`
- **AND** that observation SHALL preserve the batch-scoped `batch_count`, `known_count`, `unknown_count`, and `unvalidated_count` fields
- **AND** it SHALL aggregate that batch under the `unavailable` validator-source grouping key

### Requirement: Observe-only graph admission metrics are aggregated into runtime feedback snapshots
The runtime feedback snapshot SHALL aggregate observe-only graph admission metrics into stable metric families for downstream surfaces.

#### Scenario: Admission batch metrics are aggregated by source and validator identity
- **WHEN** graph-admission telemetry is recorded for a supported runtime graph input
- **THEN** the runtime feedback snapshot SHALL aggregate one observed batch count for that telemetry event
- **AND** event-bus admission metrics SHALL remain grouped by producer-source, producer-group, and validator-source identity
- **AND** `content_saved_extractor` admission metrics SHALL remain grouped by the synthetic source label and validator-source identity without inventing a producer-group
- **AND** the snapshot SHALL aggregate `known_count`, `unknown_count`, and `unvalidated_count` totals from those batch events

#### Scenario: Non-admission baselines preserve their counting units
- **WHEN** dropped-unknown or graph write-failure baseline telemetry is recorded
- **THEN** extractor dropped-unknown metrics SHALL aggregate one dropped-triple count per rejected triple
- **AND** graph write-failure baseline metrics SHALL aggregate one failed-batch count per failed graph write attempt

#### Scenario: Unmapped and validator identities remain visible in snapshots
- **WHEN** the runtime feedback snapshot is built
- **THEN** unmapped-source metrics SHALL remain grouped by raw source label
- **AND** validator-source groupings SHALL remain grouped by validator-source identifier, including the stable `unavailable` value when classification could not run
- **AND** both groupings SHALL use batch counts rather than triple counts

