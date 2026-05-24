## ADDED Requirements

### Requirement: Graph clear treats EOF as denial
`lango graph clear` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Graph clear EOF aborts cleanly
- **WHEN** `lango graph clear` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave the graph store unchanged
