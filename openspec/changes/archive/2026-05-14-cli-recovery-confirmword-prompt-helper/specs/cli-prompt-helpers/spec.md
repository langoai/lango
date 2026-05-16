## ADDED Requirements

### Requirement: Shared line prompt supports deterministic command streams
The shared CLI prompt package SHALL provide a visible line-entry helper that writes a prompt through a supplied output stream and reads one line from a supplied input stream without requiring process-global stdio replacement.

#### Scenario: Shared line prompt uses injected streams
- **WHEN** the visible line-entry helper is exercised in tests with injected input and output streams
- **THEN** it SHALL write the prompt to the injected output stream
- **AND** it SHALL read the entered line from the injected input stream
