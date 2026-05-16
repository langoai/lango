## ADDED Requirements

### Requirement: Change-passphrase success output routing

`lango security change-passphrase` SHALL write its non-error success confirmation through the Cobra command output stream so wrappers and test harnesses can capture completion output without intercepting process-global stdout.

#### Scenario: Change-passphrase success writes to command output
- **WHEN** `lango security change-passphrase` succeeds
- **THEN** the command writes the success confirmation to the Cobra command output stream
