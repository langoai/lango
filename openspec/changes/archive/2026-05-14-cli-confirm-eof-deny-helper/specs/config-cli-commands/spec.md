## ADDED Requirements

### Requirement: Config delete treats EOF as denial
`lango config delete <name>` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Config delete EOF aborts cleanly
- **WHEN** `lango config delete staging` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** the profile SHALL remain undeleted
