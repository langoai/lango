## ADDED Requirements

### Requirement: Safe Proxy Deployment
The Factory SHALL pass the Safe L2 singleton address (not the Safe7579 adapter) as the `_singleton` parameter to `createProxyWithNonce()`. The Safe7579 adapter SHALL be passed as the `to` parameter in `Safe.setup()` for delegate-call initialization.

#### Scenario: Deploy with correct singleton
- GIVEN a Factory initialized with both `singletonAddr` (Safe L2) and `safe7579Addr` (ERC-7579 adapter)
- WHEN `Factory.Deploy()` is called
- THEN the `createProxyWithNonce()` call uses `singletonAddr` as the singleton parameter
- AND the `Safe.setup()` initializer uses `safe7579Addr` as the delegate-call target (`to`)
- AND the deployment succeeds without "execution reverted"

#### Scenario: ComputeAddress uses correct singleton
- GIVEN a Factory with separate `singletonAddr` and `safe7579Addr`
- WHEN `Factory.ComputeAddress()` is called
- THEN the CREATE2 formula uses `singletonAddr` in the proxy initCode hash
- AND the computed address matches the on-chain deployed address

### Requirement: Safe Singleton Configuration
The system SHALL support a `SafeSingletonAddress` config field for specifying the Safe L2 singleton implementation address.

#### Scenario: Default singleton address
- GIVEN `SmartAccountConfig.SafeSingletonAddress` is empty
- WHEN the smart account subsystem initializes
- THEN the system SHALL use `0x29fcB43b46531BcA003ddC8FCB67FFE91900C762` (Safe L2 v1.4.1)

#### Scenario: Custom singleton address
- GIVEN `SmartAccountConfig.SafeSingletonAddress` is set to a custom address
- WHEN the smart account subsystem initializes
- THEN the system SHALL use the configured address as the proxy singleton

#### Scenario: TUI settings field
- GIVEN the settings TUI is opened on the smart account form
- WHEN the user views the form
- THEN a "Safe Singleton" field SHALL be visible with placeholder `0x29fcB43b46531BcA003ddC8FCB67FFE91900C762`

## ADDED Requirements

### Requirement: Bundler Revert Reason Extraction
The bundler client SHALL extract and decode revert reasons from JSON-RPC error responses.

#### Scenario: Error(string) revert
- GIVEN a bundler JSON-RPC error with `data` field containing an Error(string) ABI-encoded payload
- WHEN the error is formatted
- THEN the error message SHALL include the decoded revert string (e.g., "reason: Caller is not the owner")

#### Scenario: Panic(uint256) revert
- GIVEN a bundler JSON-RPC error with `data` field containing a Panic(uint256) payload
- WHEN the error is formatted
- THEN the error message SHALL include the panic description (e.g., "reason: panic: arithmetic overflow/underflow")

#### Scenario: Missing or unparseable data
- GIVEN a bundler JSON-RPC error without a `data` field or with undecodable data
- WHEN the error is formatted
- THEN the error message SHALL omit the revert reason (no crash, no empty reason)

### Requirement: Contract Caller Revert Diagnostics
The contract caller SHALL extract revert reasons from go-ethereum errors and replay failed transactions as eth_call to obtain revert data.

#### Scenario: go-ethereum DataError
- GIVEN a go-ethereum RPC error implementing the `dataError` interface
- WHEN `Caller.Read()` or gas estimation fails
- THEN the error message SHALL include the decoded revert reason from `ErrorData()`

#### Scenario: Receipt-level revert with eth_call replay
- GIVEN a confirmed transaction with `receipt.Status == 0` (reverted)
- WHEN `Caller.Write()` processes the receipt
- THEN the system SHALL replay the call as `eth_call` at the revert block number
- AND include the decoded revert reason in the error message

#### Scenario: Gas estimation revert fallback
- GIVEN an `EstimateGas` call that fails without a direct DataError
- WHEN `Caller.Write()` handles the estimation failure
- THEN the system SHALL replay as `eth_call` to extract the revert reason

### Requirement: Tool Execution Failure Logging
The system SHALL log tool execution failures at WARN level for server-side visibility.

#### Scenario: ADK tool handler error
- GIVEN a tool call that returns an error
- WHEN the ADK tool adapter processes the result
- THEN a WARN-level log SHALL be emitted with tool name, agent name, and error details

#### Scenario: Event bus tool failure logging
- GIVEN the observability subsystem is enabled
- WHEN a `ToolExecutedEvent` with `Success=false` is published on the event bus
- THEN the observability subscriber SHALL log the failure at WARN level with tool name, agent, session, duration, and error
