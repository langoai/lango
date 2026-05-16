## ADDED Requirements

### Requirement: Payment-send EOF aborts cleanly
`lango payment send` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Payment-send EOF aborts without submission
- **WHEN** `lango payment send` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL NOT submit a payment
