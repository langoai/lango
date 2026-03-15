## MODIFIED Requirements

### Requirement: Paymaster config structure
The `SmartAccountPaymasterConfig` SHALL include a `Mode` field (`mapstructure:"mode"`) accepting `"rpc"` (default) or `"permit"` to select the paymaster interaction mode.

#### Scenario: Mode field defaults to rpc
- **WHEN** `mode` is empty or omitted in config
- **THEN** the system SHALL treat it as `"rpc"` mode and require `rpcURL`

#### Scenario: Permit mode does not require rpcURL
- **WHEN** `mode` is `"permit"`
- **THEN** the system SHALL NOT require `rpcURL` and SHALL use the wallet provider and RPC client for permit signing
