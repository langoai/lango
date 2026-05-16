## MODIFIED Requirements

### Requirement: Recovery restore output routing
`lango security recovery restore` SHALL write its success confirmation through the Cobra command output stream so wrappers and test harnesses can capture completion output without intercepting process-global stdout.

#### Scenario: Recovery restore warning output writes to command error stream
- **WHEN** `lango security recovery restore` emits keyfile or keyring update notices or warnings
- **THEN** those messages SHALL write to the Cobra command error stream
