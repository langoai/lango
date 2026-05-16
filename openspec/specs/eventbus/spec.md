## Purpose

Capability spec for eventbus. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: AlertEvent type
The eventbus package SHALL define an AlertEvent struct with fields: Type (string), Severity (string), Message (string), Details (map[string]interface{}), SessionKey (string), and Timestamp (time.Time). The EventName() method SHALL return "alert.triggered".

#### Scenario: AlertEvent implements Event interface
- **WHEN** an AlertEvent is created
- **THEN** calling EventName() returns "alert.triggered"

### Requirement: Alert event name constant
The eventbus package SHALL define an EventAlertTriggered constant with value "alert.triggered".

#### Scenario: Constant matches EventName
- **WHEN** the EventAlertTriggered constant is used
- **THEN** its value equals the return value of AlertEvent.EventName()

### Requirement: CompactionCompletedEvent type
The eventbus package SHALL define a `CompactionCompletedEvent` struct with fields: `SessionKey string`, `UpToIndex int`, `SummaryTokens int`, `ReclaimedTokens int`, and `Timestamp time.Time`. The `EventName()` method SHALL return `"compaction.completed"`. A matching `EventCompactionCompleted` string constant SHALL be defined with the same value.

#### Scenario: Event name is stable
- **WHEN** a `CompactionCompletedEvent` is created
- **THEN** `EventName()` SHALL return `"compaction.completed"`
- **AND** the `EventCompactionCompleted` constant SHALL equal that value

### Requirement: CompactionSlowEvent type
The eventbus package SHALL define a `CompactionSlowEvent` struct with fields: `SessionKey string`, `WaitedFor time.Duration`, and `Timestamp time.Time`. It SHALL be published when the sync-point guard in `ContextAwareModelAdapter.GenerateContent()` exceeds its timeout waiting for an in-flight compaction. The `EventName()` method SHALL return `"compaction.slow"`. A matching `EventCompactionSlow` string constant SHALL be defined.

#### Scenario: Event name is stable
- **WHEN** a `CompactionSlowEvent` is created
- **THEN** `EventName()` SHALL return `"compaction.slow"`
- **AND** the `EventCompactionSlow` constant SHALL equal that value

### Requirement: LearningSuggestionEvent type
The eventbus package SHALL define a `LearningSuggestionEvent` struct with fields: `SessionKey string`, `SuggestionID string`, `Pattern string`, `ProposedRule string`, `Confidence float64`, `Rationale string`, and `Timestamp time.Time`. The `EventName()` method SHALL return `"learning.suggestion"`. A matching `EventLearningSuggestion` string constant SHALL be defined.

#### Scenario: Event name is stable
- **WHEN** a `LearningSuggestionEvent` is created
- **THEN** `EventName()` SHALL return `"learning.suggestion"`
- **AND** the `EventLearningSuggestion` constant SHALL equal that value

### Requirement: Graph admission-related event types are defined
The eventbus package SHALL define the event types required to represent observe-only graph admission telemetry and the related non-admission baseline signals for this slice.

#### Scenario: Event-bus admission event shape carries source, group, and validator fields
- **WHEN** a supported `TriplesExtractedEvent` producer-source batch is classified in observe mode
- **THEN** the graph-admission event shape SHALL include that producer-source identifier, its producer-group identifier, and the validator-source identifier used for predicate checks
- **AND** it SHALL use the stable validator-source value `ontology_registry` when classification uses the ontology service validator closure
- **AND** it SHALL include `batch_count`, `known_count`, `unknown_count`, and `unvalidated_count` fields for that observed slice

#### Scenario: Content-saved admission event shape carries synthetic source and validator fields
- **WHEN** the `content.saved` extraction path returns triples that are classified in observe mode
- **THEN** the graph-admission event shape SHALL include the synthetic `content_saved_extractor` source label and the validator-source identifier used for predicate checks
- **AND** it SHALL NOT synthesize an event-bus producer-group identifier for that telemetry
- **AND** it SHALL include `batch_count`, `known_count`, `unknown_count`, and `unvalidated_count` fields for that observed slice

#### Scenario: Validator-unavailable admission event shape carries unvalidated counts
- **WHEN** ontology is disabled or ontology initialization fails before observe-only admission can classify a batch through the shared validator closure
- **THEN** the graph-admission event shape SHALL include `known_count = 0`, `unknown_count = 0`, and `unvalidated_count = len(observed slice)`
- **AND** it SHALL still represent one observed batch for that slice
- **AND** it SHALL set the validator-source field to the stable value `unavailable`
- **AND** it SHALL include `batch_count`, `known_count`, `unknown_count`, and `unvalidated_count` fields for that observed slice

### Requirement: Non-admission baseline events remain separate telemetry families
Extractor dropped-unknown baselines, `unmapped-source` signals, and aggregate graph write-failure baselines SHALL remain separate telemetry families rather than graph-admission telemetry.

#### Scenario: Unmapped event-bus source is surfaced explicitly
- **WHEN** observe mode receives a `TriplesExtractedEvent.Source` label that is outside the supported event-bus producer-source set and therefore not assigned to a known producer group in this slice
- **THEN** the runtime SHALL record an `unmapped-source` telemetry signal carrying the original raw source label
- **AND** it SHALL NOT widen observe-only admission classification beyond the supported runtime graph inputs in this slice

#### Scenario: Extractor dropped-unknown baseline event shape stays pre-admission
- **WHEN** the `content.saved` extraction path rejects an unknown predicate before graph admission runs
- **THEN** the dropped-unknown baseline event shape SHALL include the synthetic `content_saved_extractor` source label
- **AND** it SHALL record one dropped triple per rejected triple
- **AND** it SHALL NOT imply that graph admission dropped the triple

#### Scenario: Graph write-failure baseline event shape stays aggregate
- **WHEN** a batched graph write fails while observe mode is enabled
- **THEN** the graph write-failure baseline event shape SHALL represent the failed batch as an aggregate failure signal
- **AND** it SHALL record one failed batch per failed graph write attempt
- **AND** it SHALL NOT require admission source, producer-group, or validator-source tags on that aggregate failure baseline

