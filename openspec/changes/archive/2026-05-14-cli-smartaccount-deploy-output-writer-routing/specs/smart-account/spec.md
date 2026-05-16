## ADDED Requirements

### Requirement: Smart account deploy output routing

`lango account deploy` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Smart account deploy table writes to command output
- **WHEN** `lango account deploy` is run without `--output json`
- **THEN** the command writes the human-readable deployment summary to the Cobra command output stream

#### Scenario: Smart account deploy JSON writes to command output
- **WHEN** `lango account deploy --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
