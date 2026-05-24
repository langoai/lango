## ADDED Requirements

### Requirement: Migrate-passphrase output routing

`lango security migrate-passphrase` SHALL write its non-error status and success output through the Cobra command output stream so wrappers and test harnesses can capture migration progress without intercepting process-global stdout.

#### Scenario: Migrate-passphrase progress writes to command output
- **WHEN** `lango security migrate-passphrase` runs
- **THEN** the command writes its migration guidance, progress, and success output to the Cobra command output stream
