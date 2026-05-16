## ADDED Requirements

### Requirement: Keyring-clear EOF aborts cleanly
`lango security keyring clear` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Keyring-clear EOF aborts without clearing
- **WHEN** `lango security keyring clear` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL NOT clear stored credentials
