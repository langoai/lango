## ADDED Requirements

### Requirement: P2P sandbox output routing

`lango p2p sandbox` subcommands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture status, smoke-test, and cleanup output without intercepting process-global stdout.

#### Scenario: Sandbox status output writes to command output
- **WHEN** user runs `lango p2p sandbox status`
- **THEN** the command writes the sandbox status output to the Cobra command output stream

#### Scenario: Sandbox smoke-test output writes to command output
- **WHEN** user runs `lango p2p sandbox test`
- **THEN** the command writes the runtime-selection and smoke-test result output to the Cobra command output stream

#### Scenario: Sandbox cleanup output writes to command output
- **WHEN** user runs `lango p2p sandbox cleanup`
- **THEN** the command writes the cleanup success output to the Cobra command output stream
