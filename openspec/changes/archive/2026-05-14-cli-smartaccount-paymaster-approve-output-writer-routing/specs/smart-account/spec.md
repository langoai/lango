## ADDED Requirements

### Requirement: Smart account paymaster approval output routing

`lango account paymaster approve` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Smart account paymaster approval table writes to command output
- **WHEN** `lango account paymaster approve --amount <usdc>` is run without `--output json`
- **THEN** the command writes the human-readable approval summary to the Cobra command output stream

#### Scenario: Smart account paymaster approval JSON writes to command output
- **WHEN** `lango account paymaster approve --amount <usdc> --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
