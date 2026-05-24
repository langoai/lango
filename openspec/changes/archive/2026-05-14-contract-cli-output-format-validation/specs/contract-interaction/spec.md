## ADDED Requirements
### Requirement: Contract CLI output format stays explicit and validated
`lango contract read`, `lango contract call`, and `lango contract abi load` SHALL accept `--output table|json` and reject unknown values before config loading.

#### Scenario: Contract read rejects unknown output before config load
- **WHEN** `lango contract read --address 0x... --abi ./erc20.json --method balanceOf --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Contract call rejects unknown output before config load
- **WHEN** `lango contract call --address 0x... --abi ./erc20.json --method transfer --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Contract ABI load rejects unknown output before config load
- **WHEN** `lango contract abi load --address 0x... --file ./erc20.json --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader
