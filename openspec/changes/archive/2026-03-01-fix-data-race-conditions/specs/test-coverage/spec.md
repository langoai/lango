## ADDED Requirements

### Requirement: Channel mock thread safety
Channel test mock types SHALL use mutex synchronization to protect shared slices from concurrent access by handler goroutines and test assertions.

#### Scenario: Slack mock concurrent access
- **WHEN** a slack handler goroutine appends to PostMessages/UpdateMessages while the test goroutine reads them
- **THEN** access SHALL be serialized via mutex to prevent data races

#### Scenario: Telegram mock concurrent access
- **WHEN** a telegram handler goroutine appends to SentMessages/RequestCalls while the test goroutine reads them
- **THEN** access SHALL be serialized via mutex to prevent data races

#### Scenario: Safe mock data retrieval
- **WHEN** test code reads mock recorded calls
- **THEN** helper methods SHALL return defensive copies of the underlying slices
