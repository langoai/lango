## ADDED Requirements

### Requirement: Risk CLI output routing

`lango economy risk` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled and disabled output without intercepting process-global stdout.

#### Scenario: Risk enabled output writes to command output
- **WHEN** `lango economy risk` is run with economy enabled
- **THEN** the command writes the risk configuration to the Cobra command output stream

#### Scenario: Risk disabled output writes to command output
- **WHEN** `lango economy risk` is run with economy disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream
