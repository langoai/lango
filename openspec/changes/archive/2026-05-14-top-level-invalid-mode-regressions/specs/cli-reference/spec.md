## MODIFIED Requirements

### Requirement: TUI startup notices remain seam-aware
Interactive top-level TUI entrypoints SHALL route their startup notice text through seam-aware stderr writers so wrapper and regression captures do not depend on process-global stderr interception.

#### Scenario: Top-level interactive entrypoints reject unknown modes
- **WHEN** bare `lango`, `lango cockpit`, or `lango chat` is started with an unknown `--mode`
- **THEN** the command SHALL return an actionable `unknown mode` error
