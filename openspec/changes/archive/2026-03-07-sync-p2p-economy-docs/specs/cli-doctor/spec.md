## ADDED Requirements

### Requirement: Economy health check
The doctor command SHALL include an EconomyCheck that validates economy layer configuration. The check SHALL skip when `economy.enabled` is false. When enabled, it SHALL validate that `budget.defaultMax` is parseable as a float, `risk.highTrustScore > risk.mediumTrustScore`, `escrow.maxMilestones > 0`, `negotiate.maxRounds > 0`, and `pricing.minPrice` is parseable as a float.

#### Scenario: Economy disabled
- **WHEN** doctor runs with `economy.enabled` set to false
- **THEN** EconomyCheck returns StatusSkip with message "Economy layer is disabled"

#### Scenario: Valid economy config
- **WHEN** economy is enabled with valid budget, risk, escrow, negotiation, and pricing settings
- **THEN** EconomyCheck returns StatusPass

#### Scenario: Invalid budget defaultMax
- **WHEN** economy is enabled and `budget.defaultMax` cannot be parsed as a float
- **THEN** EconomyCheck returns StatusFail with message identifying the parse error

#### Scenario: Risk score ordering
- **WHEN** economy is enabled and `risk.highTrustScore <= risk.mediumTrustScore`
- **THEN** EconomyCheck returns StatusWarn indicating high trust score should exceed medium trust score

### Requirement: Contract health check
The doctor command SHALL include a ContractCheck that validates contract interaction prerequisites. The check SHALL skip when `payment.enabled` is false. When enabled, it SHALL validate that `payment.network.rpcURL` and `payment.network.chainID` are set.

#### Scenario: Payment disabled
- **WHEN** doctor runs with `payment.enabled` set to false
- **THEN** ContractCheck returns StatusSkip with message "Payment/contract is disabled"

#### Scenario: Missing RPC URL
- **WHEN** payment is enabled but `payment.network.rpcURL` is empty
- **THEN** ContractCheck returns StatusFail with message indicating RPC URL is required

#### Scenario: Valid contract config
- **WHEN** payment is enabled with rpcURL and chainID set
- **THEN** ContractCheck returns StatusPass

### Requirement: Observability health check
The doctor command SHALL include an ObservabilityCheck that validates observability configuration. The check SHALL skip when `observability.enabled` is false. When enabled, it SHALL validate that `tokens.retentionDays > 0` when `persistHistory` is true, `health.interval > 0`, and `audit.retentionDays > 0`.

#### Scenario: Observability disabled
- **WHEN** doctor runs with `observability.enabled` set to false
- **THEN** ObservabilityCheck returns StatusSkip with message "Observability is disabled"

#### Scenario: Invalid retention days
- **WHEN** observability is enabled with `tokens.persistHistory` true and `tokens.retentionDays` is 0
- **THEN** ObservabilityCheck returns StatusWarn indicating retention days should be positive

#### Scenario: Valid observability config
- **WHEN** observability is enabled with valid token, health, and audit settings
- **THEN** ObservabilityCheck returns StatusPass

### Requirement: New checks registered in AllChecks
The EconomyCheck, ContractCheck, and ObservabilityCheck SHALL be registered in the `AllChecks()` function so they are executed by the `lango doctor` command.

#### Scenario: Doctor runs economy, contract, and observability checks
- **WHEN** user runs `lango doctor`
- **THEN** the output includes results for "Economy Layer", "Smart Contracts", and "Observability" checks
