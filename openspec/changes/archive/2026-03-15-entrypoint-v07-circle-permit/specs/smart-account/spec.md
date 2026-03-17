## MODIFIED Requirements

### Requirement: Bundler RPC field format
The bundler client `userOpToMap()` SHALL emit v0.7 split fields for UserOperation serialization.

#### Scenario: InitCode splitting
- **WHEN** `initCode` is 20 bytes or longer
- **THEN** the system SHALL split into `factory` (first 20 bytes as address) and `factoryData` (remaining bytes as hex)

#### Scenario: Empty initCode
- **WHEN** `initCode` is empty or shorter than 20 bytes
- **THEN** the system SHALL emit `factory: "0x"` and `factoryData: "0x"`

#### Scenario: PaymasterAndData splitting
- **WHEN** `paymasterAndData` is 52 bytes or longer
- **THEN** the system SHALL split into `paymaster` (first 20 bytes as address), `paymasterVerificationGasLimit` (next 16 bytes as uint128 hex), `paymasterPostOpGasLimit` (next 16 bytes as uint128 hex), and `paymasterData` (remaining bytes as hex)

#### Scenario: Empty paymasterAndData
- **WHEN** `paymasterAndData` is empty or shorter than 52 bytes
- **THEN** the system SHALL emit `paymaster: "0x"`, `paymasterVerificationGasLimit: "0x0"`, `paymasterPostOpGasLimit: "0x0"`, `paymasterData: "0x"`

### Requirement: Gas estimation paymaster fields
The `GasEstimate` struct SHALL include optional `PaymasterVerificationGasLimit` and `PaymasterPostOpGasLimit` fields for v0.7 bundler responses.

#### Scenario: Bundler returns paymaster gas fields
- **WHEN** the bundler response includes `paymasterVerificationGasLimit` and `paymasterPostOpGasLimit`
- **THEN** the system SHALL parse and populate these fields in `GasEstimate`

#### Scenario: Bundler omits paymaster gas fields
- **WHEN** the bundler response does not include paymaster gas fields
- **THEN** the `PaymasterVerificationGasLimit` and `PaymasterPostOpGasLimit` fields SHALL be nil
