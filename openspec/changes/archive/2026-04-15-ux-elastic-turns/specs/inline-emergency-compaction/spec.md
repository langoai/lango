## ADDED Requirements

### Requirement: SessionCompactor interface
The system SHALL define a `SessionCompactor` interface with method `CompactMessages(key string, upToIndex int, summary string) error`. This interface SHALL be injectable into `ContextAwareModelAdapter` via a `WithSessionCompactor()` method following the existing `WithMemory()` pattern.

#### Scenario: Compactor injection
- **WHEN** `WithSessionCompactor(compactor)` is called on a `ContextAwareModelAdapter`
- **THEN** the adapter SHALL store the compactor for use during `GenerateContent()`

#### Scenario: Nil compactor
- **WHEN** `WithSessionCompactor` is not called
- **THEN** the adapter SHALL NOT attempt emergency compaction (compactor is nil)

### Requirement: Emergency compaction trigger in GenerateContent
After Phase 2 (budget measurement) in `GenerateContent()`, the adapter SHALL check if `measured total tokens > modelWindow × 0.9`. If true and a `SessionCompactor` is available, the adapter SHALL invoke `CompactMessages()` synchronously, preserving the first 3 and last 6 messages, then restart from Phase 1 (retrieval). The compaction SHALL execute at most once per `GenerateContent()` call.

#### Scenario: Compaction triggered at 90% threshold
- **WHEN** measured total tokens exceed 90% of the model window
- **AND** a `SessionCompactor` is injected
- **THEN** the adapter SHALL call `CompactMessages()` with the session key
- **AND** SHALL restart retrieval from Phase 1 with the compacted message set

#### Scenario: Below threshold no compaction
- **WHEN** measured total tokens are below 90% of the model window
- **THEN** no compaction SHALL occur

#### Scenario: Compaction runs at most once
- **WHEN** compaction is triggered and the re-measured total still exceeds 90%
- **THEN** the adapter SHALL NOT trigger a second compaction
- **AND** SHALL proceed with the over-budget context (best effort)

### Requirement: budgets.Degraded is not a compaction trigger
The `budgets.Degraded` flag SHALL NOT trigger emergency compaction. `Degraded` indicates that the base prompt alone exceeds the model window budget, which is a configuration issue that session message compaction cannot resolve. When `Degraded` is true, the adapter SHALL log a warning and emit a user-facing message indicating the model window is too small for the current configuration.

#### Scenario: Degraded does not trigger compaction
- **WHEN** `budgets.Degraded` is true
- **THEN** the adapter SHALL NOT invoke `CompactMessages()`
- **AND** SHALL log a warning about model window being too small for base prompt

### Requirement: SessionCompactor wiring
The `app/wiring.go` module SHALL inject the session `EntStore` as a `SessionCompactor` into the `ContextAwareModelAdapter` during application initialization.

#### Scenario: Compactor wired at startup
- **WHEN** the application initializes and creates a `ContextAwareModelAdapter`
- **THEN** the adapter SHALL receive the `EntStore` as its `SessionCompactor`
