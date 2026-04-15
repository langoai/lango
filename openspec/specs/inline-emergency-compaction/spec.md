# inline-emergency-compaction Specification

## Purpose
TBD - created by archiving change ux-elastic-turns. Update Purpose after archive.
## Requirements
### Requirement: SessionCompactor interface
The system SHALL define a `SessionCompactor` interface with method `CompactMessages(key string, upToIndex int, summary string) error`. This interface SHALL be injectable into `ContextAwareModelAdapter` via a `WithSessionCompactor()` method following the existing `WithMemory()` pattern.

#### Scenario: Compactor injection
- **WHEN** `WithSessionCompactor(compactor)` is called on a `ContextAwareModelAdapter`
- **THEN** the adapter SHALL store the compactor for use during `GenerateContent()`

#### Scenario: Nil compactor
- **WHEN** `WithSessionCompactor` is not called
- **THEN** the adapter SHALL NOT attempt emergency compaction (compactor is nil)

### Requirement: Emergency compaction trigger measurement
The emergency compaction trigger SHALL measure the TOTAL context size including conversation history (`req.Contents`), base prompt tokens, and all injected sections (Knowledge, RAG, Memory, RunSummary). The threshold comparison SHALL use `modelWindow × 0.9`.

#### Scenario: Long conversation triggers compaction
- **WHEN** a session has 100K tokens of conversation history, 8K base prompt, and 5K injected context
- **AND** the model window is 128K tokens
- **THEN** `totalMeasured` SHALL be approximately 113K (100K + 8K + 5K)
- **AND** compaction SHALL trigger because 113K > 115.2K × 0.9

#### Scenario: Short conversation with heavy context does not over-trigger
- **WHEN** a session has 5K tokens of conversation history, 8K base prompt, and 20K injected context
- **AND** the model window is 128K tokens
- **THEN** `totalMeasured` SHALL be approximately 33K
- **AND** compaction SHALL NOT trigger because 33K < 115.2K

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

