## ADDED Requirements

### Requirement: Workflow management output routing

`lango workflow list`, `status`, `history`, and `cancel` SHALL write non-error output through the Cobra command output stream so wrappers and test harnesses can capture table, detail, empty-state, and confirmation output without intercepting process-global stdout.

#### Scenario: Workflow list writes to command output
- **WHEN** user runs `lango workflow list`
- **THEN** the list table or empty-state message writes to the Cobra command output stream

#### Scenario: Workflow status writes to command output
- **WHEN** user runs `lango workflow status <uuid>`
- **THEN** the status detail output writes to the Cobra command output stream

#### Scenario: Workflow history writes to command output
- **WHEN** user runs `lango workflow history`
- **THEN** the history table or empty-state message writes to the Cobra command output stream

#### Scenario: Workflow cancel writes to command output
- **WHEN** user runs `lango workflow cancel <uuid>`
- **THEN** the cancellation confirmation writes to the Cobra command output stream
