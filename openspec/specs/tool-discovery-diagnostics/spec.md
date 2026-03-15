# tool-discovery-diagnostics Specification

## Purpose
TBD - created by archiving change tool-discovery-audit-bugfix. Update Purpose after archive.
## Requirements
### Requirement: Config validation for SmartAccount
SmartAccountConfig SHALL provide a Validate() method that returns an error when enabled but required fields (EntryPointAddress, FactoryAddress, BundlerURL) are empty. Validate() SHALL return nil when the config is disabled.

#### Scenario: Enabled with missing entryPointAddress
- **WHEN** SmartAccountConfig has Enabled=true and EntryPointAddress=""
- **THEN** Validate() returns error containing "smartAccount.entryPointAddress is required"

#### Scenario: Enabled with missing factoryAddress
- **WHEN** SmartAccountConfig has Enabled=true and FactoryAddress=""
- **THEN** Validate() returns error containing "smartAccount.factoryAddress is required"

#### Scenario: Enabled with missing bundlerURL
- **WHEN** SmartAccountConfig has Enabled=true and BundlerURL=""
- **THEN** Validate() returns error containing "smartAccount.bundlerURL is required"

#### Scenario: Disabled config is always valid
- **WHEN** SmartAccountConfig has Enabled=false
- **THEN** Validate() returns nil regardless of other field values

### Requirement: Payment RPCURL pre-validation
The payment wiring SHALL validate that RPCURL is non-empty before calling ethclient.Dial, and SHALL log a warning with actionable fix instructions when empty.

#### Scenario: Empty RPCURL skips payment init
- **WHEN** payment is enabled but RPCURL is ""
- **THEN** initPayment logs warning with "payment RPC URL not configured" and returns nil

### Requirement: Warning logs for degraded subsystem wiring
The smart account wiring SHALL emit warning logs when risk engine or sentinel guard cannot be wired due to missing dependencies. X402 wiring SHALL warn when secrets are nil. Observability SHALL warn when disabled but sub-features are enabled.

#### Scenario: Risk engine not available
- **WHEN** smart account is enabled but economy components are nil
- **THEN** wiring logs warning "smart account: risk engine not wired, spending controls unavailable"

#### Scenario: Sentinel guard not available
- **WHEN** smart account is enabled but sentinel engine or bus is nil
- **THEN** wiring logs warning "smart account: sentinel session guard not wired, anomaly detection unavailable"

#### Scenario: X402 secrets nil
- **WHEN** payment is enabled but secrets store is nil
- **THEN** X402 init logs warning "X402 interceptor requires security.signer, skipping"

#### Scenario: Observability sub-flag conflict
- **WHEN** observability.enabled is false but tokens.enabled or health.enabled is true
- **THEN** wiring logs warning "observability disabled but sub-features enabled; sub-features ignored"

### Requirement: SessionGuard lifecycle management
SessionGuard SHALL provide a Stop() method that deactivates alert processing. After Stop() is called, handleAlert SHALL return without executing revoke or restrict callbacks.

#### Scenario: Stop disables alert handling
- **WHEN** SessionGuard.Start() has been called, then Stop() is called
- **THEN** subsequent SentinelAlertEvent publications do not trigger revoke or restrict callbacks

#### Scenario: SessionGuard registered with lifecycle registry
- **WHEN** sentinel session guard is wired in initSmartAccount
- **THEN** the guard is registered with the lifecycle registry at PriorityAutomation for graceful shutdown

