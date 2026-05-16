## ADDED Requirements

### Requirement: Shared raw line reader preserves caller control
The system SHALL provide a shared lower-level helper that reads one raw line from a supplied `io.Reader` and returns the line text and read error without performing caller-specific normalization.

#### Scenario: Shared line reader returns a full line with newline
- **WHEN** the helper reads input ending in `\n`
- **THEN** it SHALL return the raw line including the trailing newline

#### Scenario: Shared line reader returns partial line on EOF
- **WHEN** the helper encounters EOF after reading partial line content
- **THEN** it SHALL return the partial line together with the EOF error so the caller can decide how to handle it
