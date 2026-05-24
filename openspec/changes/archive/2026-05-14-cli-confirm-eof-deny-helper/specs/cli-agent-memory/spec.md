## ADDED Requirements

### Requirement: Memory clear treats EOF as denial
`lango memory clear <session-key>` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Memory clear EOF aborts cleanly
- **WHEN** `lango memory clear user-123` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave memory entries untouched
