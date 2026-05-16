## MODIFIED Requirements

### Requirement: TUI startup notices remain seam-aware
Interactive top-level TUI entrypoints SHALL route their startup notice text through seam-aware stderr writers so wrapper and regression captures do not depend on process-global stderr interception.

#### Scenario: Chat startup notices write through seam-aware stderr
- **WHEN** `lango chat` begins startup
- **THEN** the banner, log-path notice, and initializing line SHALL be written through the chat stderr seam
