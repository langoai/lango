## MODIFIED Requirements

### Requirement: Paymaster CLI commands
The system SHALL provide `lango account paymaster status` and `lango account paymaster approve` commands.

#### Scenario: CLI status output includes mode
- **WHEN** `lango account paymaster status` is run
- **THEN** it SHALL display the `mode` field (`rpc` or `permit`) in both table and JSON output

#### Scenario: CLI status omits RPC URL in permit mode
- **WHEN** paymaster mode is `permit` and no RPC URL is configured
- **THEN** the status output SHALL omit the RPC URL line in table format

### Requirement: Paymaster configuration
The system SHALL support `SmartAccountPaymasterConfig` with enabled, provider, mode, rpcURL, tokenAddress, paymasterAddress, and policyId fields. The `mode` field accepts `"rpc"` (default) or `"permit"`.

#### Scenario: TUI form includes mode selection
- **WHEN** the user opens the SA Paymaster Configuration form
- **THEN** a Mode select field SHALL be available with options `rpc` and `permit`

#### Scenario: TUI state update handles mode
- **WHEN** the user changes the Mode field in the TUI form
- **THEN** the config state SHALL be updated with the selected mode value
