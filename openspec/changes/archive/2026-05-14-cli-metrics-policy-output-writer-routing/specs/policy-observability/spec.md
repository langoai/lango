## ADDED Requirements

### Requirement: CLI policy metrics output routing

`lango metrics policy` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Policy table output writes to command output
- **WHEN** `lango metrics policy` is run without --output flag
- **THEN** the command writes the table summary to the Cobra command output stream

#### Scenario: Policy JSON output writes to command output
- **WHEN** `lango metrics policy --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
