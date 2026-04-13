# Paymaster Specification

## Purpose

Capability spec for paymaster. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: PaymasterProvider interface
The system SHALL define a `PaymasterProvider` interface with `SponsorUserOp(ctx, req) (result, error)` and `Type() string` methods for paymaster integration.

#### Scenario: Provider implements interface
- **WHEN** a Circle, Pimlico, or Alchemy provider is created
- **THEN** it SHALL implement the `PaymasterProvider` interface

### Requirement: Circle Paymaster provider
The system SHALL support Circle Paymaster via `pm_sponsorUserOperation` JSON-RPC endpoint.

#### Scenario: Successful sponsorship
- **WHEN** Circle provider receives a valid SponsorRequest
- **THEN** it SHALL return PaymasterAndData bytes from the RPC response

#### Scenario: RPC error
- **WHEN** Circle provider receives an RPC error response
- **THEN** it SHALL return an error wrapping `ErrPaymasterRejected`

#### Scenario: Optional gas overrides
- **WHEN** the RPC response includes callGasLimit, verificationGasLimit, or preVerificationGas
- **THEN** the provider SHALL parse and include them in `SponsorResult.GasOverrides`

### Requirement: Pimlico Paymaster provider
The system SHALL support Pimlico Paymaster via `pm_sponsorUserOperation` with optional `sponsorshipPolicyId`.

#### Scenario: Sponsorship with policy ID
- **WHEN** a policy ID is configured
- **THEN** the provider SHALL include it as the third parameter in the RPC call

### Requirement: Alchemy Paymaster provider
The system SHALL support Alchemy Gas Manager via `alchemy_requestGasAndPaymasterAndData` combined endpoint.

#### Scenario: Combined gas and paymaster data
- **WHEN** Alchemy provider sponsors a UserOp
- **THEN** it SHALL return both paymasterAndData and gas overrides in a single response

### Requirement: Two-phase paymaster flow in Manager
The `Manager.submitUserOp()` SHALL support a two-phase paymaster interaction: stub phase for gas estimation and final phase for signed data.

#### Scenario: Stub phase provides data for gas estimation
- **WHEN** `paymasterFn` is set and `submitUserOp` is called
- **THEN** it SHALL call `paymasterFn(ctx, op, true)` before gas estimation and set `op.PaymasterAndData` to the stub data

#### Scenario: Final phase provides signed data after gas estimation
- **WHEN** gas estimation completes
- **THEN** it SHALL call `paymasterFn(ctx, op, false)` and apply the final paymasterAndData and any gas overrides

#### Scenario: No paymaster configured
- **WHEN** `paymasterFn` is nil
- **THEN** the existing non-paymaster flow SHALL execute unchanged

#### Scenario: Stub phase failure
- **WHEN** the stub phase returns an error
- **THEN** `submitUserOp` SHALL return the error without proceeding to gas estimation

#### Scenario: Final phase failure
- **WHEN** the final phase returns an error
- **THEN** `submitUserOp` SHALL return the error without proceeding to signing

### Requirement: Gas overrides application
When `PaymasterGasOverrides` contains non-nil values, the Manager SHALL use them to override the bundler's gas estimates.

#### Scenario: Partial gas override
- **WHEN** only `CallGasLimit` is set in overrides
- **THEN** only `CallGasLimit` SHALL be overridden; other gas values remain from the bundler estimate

### Requirement: USDC approval helper
The system SHALL provide `BuildApproveCalldata(spender, amount)` and `NewApprovalCall(token, paymaster, amount)` for ERC-20 approve calldata generation.

#### Scenario: Approve calldata format
- **WHEN** `BuildApproveCalldata` is called
- **THEN** it SHALL return 68 bytes: 4-byte selector `0x095ea7b3` + 32-byte address + 32-byte amount

### Requirement: Paymaster configuration
The system SHALL support `SmartAccountPaymasterConfig` with enabled, provider, mode, rpcURL, tokenAddress, paymasterAddress, and policyId fields. The `mode` field accepts `"rpc"` (default) or `"permit"`.

#### Scenario: Provider selection
- **WHEN** config specifies provider as "circle", "pimlico", or "alchemy"
- **THEN** the corresponding provider SHALL be initialized during app wiring

#### Scenario: Mode field defaults to rpc
- **WHEN** `mode` is empty or omitted in config
- **THEN** the system SHALL treat it as `"rpc"` mode and require `rpcURL`

#### Scenario: Permit mode does not require rpcURL
- **WHEN** `mode` is `"permit"`
- **THEN** the system SHALL NOT require `rpcURL` and SHALL use the wallet provider and RPC client for permit signing

### Requirement: Paymaster agent tools
The system SHALL provide `paymaster_status` (Safe) and `paymaster_approve` (Dangerous) agent tools.

#### Scenario: Status check
- **WHEN** `paymaster_status` is called
- **THEN** it SHALL return whether paymaster is enabled and which provider is configured

#### Scenario: USDC approval
- **WHEN** `paymaster_approve` is called with token, paymaster, and amount
- **THEN** it SHALL execute an ERC-20 approve transaction via the smart account

### Requirement: Paymaster CLI commands
The system SHALL provide `lango account paymaster status` and `lango account paymaster approve` commands.

#### Scenario: CLI status output
- **WHEN** `lango account paymaster status` is run
- **THEN** it SHALL display paymaster configuration in table or JSON format

#### Scenario: CLI status output includes mode
- **WHEN** `lango account paymaster status` is run
- **THEN** it SHALL display the `mode` field (`rpc` or `permit`) in both table and JSON output

#### Scenario: CLI status omits RPC URL in permit mode
- **WHEN** paymaster mode is `permit` and no RPC URL is configured
- **THEN** the status output SHALL omit the RPC URL line in table format

#### Scenario: CLI approve with amount flag
- **WHEN** `lango account paymaster approve --amount 1000.00` is run
- **THEN** it SHALL show the approval details and instruct to use the agent tool for execution

#### Scenario: TUI form includes mode selection
- **WHEN** the user opens the SA Paymaster Configuration form
- **THEN** a Mode select field SHALL be available with options `rpc` and `permit`

#### Scenario: TUI state update handles mode
- **WHEN** the user changes the Mode field in the TUI form
- **THEN** the config state SHALL be updated with the selected mode value
