## ADDED Requirements

### Requirement: Workflow run output routing

`lango workflow run` SHALL write its validation, schedule guidance, and direct execution status output through the Cobra command output stream so wrappers and test harnesses can capture non-error output without intercepting process-global stdout.

#### Scenario: Workflow run schedule guidance writes to command output
- **WHEN** user runs `lango workflow run report.flow.yaml --schedule "0 9 * * MON"`
- **THEN** the validated workflow summary and schedule guidance write to the Cobra command output stream
