## ADDED Requirements

### Requirement: Smart account policy output routing

`lango account policy show` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table, JSON, and empty-policy output without intercepting process-global stdout.

#### Scenario: Smart account policy show table writes to command output
- **WHEN** `lango account policy show` is run without `--output json`
- **THEN** the command writes the human-readable policy summary to the Cobra command output stream

#### Scenario: Smart account policy show JSON writes to command output
- **WHEN** `lango account policy show --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream

### Requirement: Smart account policy set output routing

`lango account policy set` SHALL write the update summary through the Cobra command output stream so wrappers and test harnesses can capture confirmation output without intercepting process-global stdout.

#### Scenario: Smart account policy set success writes to command output
- **WHEN** `lango account policy set` succeeds
- **THEN** the command writes the update summary to the Cobra command output stream
