## ADDED Requirements

### Requirement: Smart account info output routing

`lango account info` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Smart account info table writes to command output
- **WHEN** `lango account info` is run without `--output json`
- **THEN** the command writes the human-readable account summary to the Cobra command output stream

#### Scenario: Smart account info JSON writes to command output
- **WHEN** `lango account info --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
