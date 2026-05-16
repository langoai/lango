## ADDED Requirements

### Requirement: Negotiate CLI output routing

`lango economy negotiate` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled and disabled output without intercepting process-global stdout.

#### Scenario: Negotiate enabled output writes to command output
- **WHEN** `lango economy negotiate` is run with negotiation enabled
- **THEN** the command writes the negotiation configuration to the Cobra command output stream

#### Scenario: Negotiate disabled output writes to command output
- **WHEN** `lango economy negotiate` is run with negotiation disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream
