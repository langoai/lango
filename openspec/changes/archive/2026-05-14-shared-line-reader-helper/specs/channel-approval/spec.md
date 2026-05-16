## ADDED Requirements

### Requirement: TTY approval fallback uses shared raw line reader
The TTY approval fallback SHALL use the shared raw line reader for its `[y/a/N]` prompt input path.

#### Scenario: TTY approval reads line through shared helper
- **WHEN** `TTYProvider.RequestApproval` reads the operator response
- **THEN** the raw line input SHALL be obtained through the shared lower-level line reader
