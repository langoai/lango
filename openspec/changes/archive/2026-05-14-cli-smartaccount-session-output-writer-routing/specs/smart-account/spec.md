## ADDED Requirements

### Requirement: Smart account session output routing

`lango account session create` and `lango account session list` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table, JSON, and empty-state output without intercepting process-global stdout.

#### Scenario: Smart account session create table writes to command output
- **WHEN** `lango account session create` is run without `--output json`
- **THEN** the command writes the human-readable session summary to the Cobra command output stream

#### Scenario: Smart account session create JSON writes to command output
- **WHEN** `lango account session create --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Smart account session list table writes to command output
- **WHEN** `lango account session list` is run without `--output json`
- **THEN** the command writes the session table to the Cobra command output stream

#### Scenario: Smart account session list JSON writes to command output
- **WHEN** `lango account session list --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Smart account session list empty-state writes to command output
- **WHEN** `lango account session list` is run with no session keys
- **THEN** the command writes the empty-state message to the Cobra command output stream

### Requirement: Smart account session revoke output routing

`lango account session revoke` SHALL write revocation confirmation through the Cobra command output stream so wrappers and test harnesses can capture success output without intercepting process-global stdout.

#### Scenario: Smart account session revoke single writes to command output
- **WHEN** `lango account session revoke <session-id>` succeeds
- **THEN** the command writes the single-session revocation confirmation to the Cobra command output stream

#### Scenario: Smart account session revoke all writes to command output
- **WHEN** `lango account session revoke --all` succeeds
- **THEN** the command writes the bulk revocation confirmation to the Cobra command output stream
