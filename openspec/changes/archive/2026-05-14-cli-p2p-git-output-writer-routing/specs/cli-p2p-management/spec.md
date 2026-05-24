## ADDED Requirements

### Requirement: P2P git output routing

`lango p2p git` guidance commands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON guidance output without intercepting process-global stdout.

#### Scenario: Git guidance text writes to command output
- **WHEN** user runs `lango p2p git init`, `log`, `diff`, `push`, or `fetch`
- **THEN** the command writes its guidance text to the Cobra command output stream

#### Scenario: Git log JSON writes to command output
- **WHEN** user runs `lango p2p git log <workspace-id> --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream
