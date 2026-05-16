## ADDED Requirements

### Requirement: Secrets-delete EOF aborts cleanly
`lango security secrets delete <name>` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Secrets-delete EOF aborts without deletion
- **WHEN** `lango security secrets delete api-key` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave the secret undeleted
