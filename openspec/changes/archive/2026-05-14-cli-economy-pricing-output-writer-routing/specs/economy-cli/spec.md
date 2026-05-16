## ADDED Requirements

### Requirement: Pricing CLI output routing

`lango economy pricing` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled and disabled output without intercepting process-global stdout.

#### Scenario: Pricing enabled output writes to command output
- **WHEN** `lango economy pricing` is run with pricing enabled
- **THEN** the command writes the pricing configuration to the Cobra command output stream

#### Scenario: Pricing disabled output writes to command output
- **WHEN** `lango economy pricing` is run with pricing disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream
