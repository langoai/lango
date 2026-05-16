## ADDED Requirements

### Requirement: Smart account paymaster status output routing

`lango account paymaster status` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Smart account paymaster status table writes to command output
- **WHEN** `lango account paymaster status` is run without `--output json`
- **THEN** the command writes the human-readable paymaster summary to the Cobra command output stream

#### Scenario: Smart account paymaster status JSON writes to command output
- **WHEN** `lango account paymaster status --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
