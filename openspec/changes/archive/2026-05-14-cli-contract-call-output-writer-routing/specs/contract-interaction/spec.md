## ADDED Requirements

### Requirement: Contract call output routing

`lango contract call` SHALL write validation payloads through the Cobra command output stream and informational runtime notes through the Cobra command error stream so wrappers and test harnesses can capture them without intercepting process-global stdout/stderr.

#### Scenario: CLI call text output writes to command streams
- **WHEN** `lango contract call --address 0x... --abi ./erc20.json --method transfer` is run
- **THEN** the command writes the validation summary to the Cobra command output stream
- **AND** it writes the runtime note to the Cobra command error stream

#### Scenario: CLI call JSON output writes to command streams
- **WHEN** `lango contract call --address 0x... --abi ./erc20.json --method transfer --output` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
- **AND** it writes the runtime note to the Cobra command error stream
