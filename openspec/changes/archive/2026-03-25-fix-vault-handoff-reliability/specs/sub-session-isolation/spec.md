## ADDED Requirements

### Requirement: Streaming failure paths discard isolated child sessions
Streaming runtime failure paths SHALL discard active isolated child sessions using the same classified discard behavior as collection-based failure paths.

#### Scenario: Iterator error discards active isolated child
- **WHEN** `RunStreamingDetailed` receives an iterator error while an isolated specialist child session is active
- **THEN** the runtime SHALL discard the active child session with the classified discard reason before returning the error
- **AND** the next retry or turn SHALL not observe the stale child overlay

#### Scenario: Streaming discard keeps parent persistence compact
- **WHEN** a streaming failure discards an isolated child session
- **THEN** persisted parent history SHALL include at most the compact root-authored discard note
- **AND** raw isolated specialist assistant and tool messages SHALL remain absent from the parent store
