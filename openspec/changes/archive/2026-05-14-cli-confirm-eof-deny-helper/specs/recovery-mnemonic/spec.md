## ADDED Requirements

### Requirement: Recovery written-down confirmation treats EOF as denial
`lango security recovery setup` SHALL treat EOF on the written-down confirmation prompt as a clean setup abort.

#### Scenario: Recovery setup EOF aborts before word checks
- **WHEN** the written-down confirmation prompt reaches EOF before approval
- **THEN** the command SHALL abort setup
- **AND** it SHALL not proceed to the confirmation-word prompts
