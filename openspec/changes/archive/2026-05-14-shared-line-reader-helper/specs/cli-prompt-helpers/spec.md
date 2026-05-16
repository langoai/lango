## ADDED Requirements

### Requirement: CLI prompt helpers use shared raw line reader
The shared CLI prompt package SHALL build its visible line-entry prompt helper on top of the shared raw line reader instead of owning a second local line-reader implementation.

#### Scenario: CLI prompt helper delegates raw line reading
- **WHEN** `ReadLineIO(...)` reads from an injected stream
- **THEN** the visible prompt helper SHALL delegate raw line reading to the shared lower-level helper
