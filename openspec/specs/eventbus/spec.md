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

