## ADDED Requirements

### Requirement: Smart account module list output routing

`lango account module list` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table, JSON, and empty-state output without intercepting process-global stdout.

#### Scenario: Smart account module list table writes to command output
- **WHEN** `lango account module list` is run without `--output json`
- **THEN** the command writes the installed modules table to the Cobra command output stream

#### Scenario: Smart account module list JSON writes to command output
- **WHEN** `lango account module list --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Smart account module list empty-state writes to command output
- **WHEN** `lango account module list` is run with no registered modules
- **THEN** the command writes the empty-state message to the Cobra command output stream
