## ADDED Requirements

### Requirement: TUI startup notices remain seam-aware
Interactive top-level TUI entrypoints SHALL route their startup notice text through seam-aware stderr writers so wrapper and regression captures do not depend on process-global stderr interception.

#### Scenario: Cockpit startup notices write through seam-aware stderr
- **WHEN** `lango cockpit` begins startup
- **THEN** the banner, log-path notice, and initializing line SHALL be written through the cockpit stderr seam

#### Scenario: Workbench startup notices write through seam-aware stderr
- **WHEN** bare `lango` begins workbench startup
- **THEN** the banner, log-path notice, and initializing line SHALL be written through the workbench stderr seam
