## ADDED Requirements

### Requirement: Contract ABI load output routing

`lango contract abi load` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON output without intercepting process-global stdout.

#### Scenario: CLI abi load text output writes to command output
- **WHEN** `lango contract abi load --address 0x... --file ./erc20.json` is run
- **THEN** the command writes the text summary to the Cobra command output stream

#### Scenario: CLI abi load JSON output writes to command output
- **WHEN** `lango contract abi load --address 0x... --file ./erc20.json --output` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
