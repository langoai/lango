## ADDED Requirements

### Requirement: lango account exec guard
The `blockLangoExec` function SHALL include a guard entry for `lango account` that redirects the agent to the built-in smart account tools.

#### Scenario: Agent attempts lango account CLI
- **WHEN** the agent attempts to run `lango account deploy` or any `lango account` subcommand via exec
- **THEN** `blockLangoExec` SHALL return a message listing all smart account tool names
- **AND** the message SHALL instruct the agent to use built-in tools instead

### Requirement: Init logging with config hints
The `initSmartAccount()` function SHALL include actionable configuration hints in its log messages when initialization is skipped.

#### Scenario: Smart account disabled
- **WHEN** `cfg.SmartAccount.Enabled` is false
- **THEN** the log message SHALL include a "fix" field with the command to enable it

#### Scenario: Payment components missing
- **WHEN** payment components are nil
- **THEN** the log message SHALL include a "fix" field listing required payment config keys
