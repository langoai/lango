## MODIFIED Requirements

### Requirement: Change-passphrase success output routing
`lango security change-passphrase` SHALL write its non-error success confirmation through the Cobra command output stream so wrappers and test harnesses can capture completion output without intercepting process-global stdout.

#### Scenario: Change-passphrase warning output writes to command error stream
- **WHEN** `lango security change-passphrase` emits keyfile or keyring update notices or warnings
- **THEN** those messages SHALL write to the Cobra command error stream
