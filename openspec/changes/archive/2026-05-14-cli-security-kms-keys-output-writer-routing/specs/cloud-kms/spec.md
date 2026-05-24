## ADDED Requirements

### Requirement: KMS keys output routing

`lango security kms keys` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, and JSON output without intercepting process-global stdout.

#### Scenario: Empty KMS key registry writes to command output
- **WHEN** `lango security kms keys` is run and the KeyRegistry is empty
- **THEN** the command writes `No keys registered.` to the Cobra command output stream

#### Scenario: Tabular KMS key listing writes to command output
- **WHEN** `lango security kms keys` is run with registered keys and `--json` is not set
- **THEN** the command writes the tabular listing to the Cobra command output stream

#### Scenario: JSON KMS key listing writes to command output
- **WHEN** `lango security kms keys --json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
