## ADDED Requirements

### Requirement: Observational Memory diagnostic check
The system SHALL include an ObservationalMemoryCheck in the doctor command that validates OM configuration. The check SHALL skip when `observationalMemory.enabled` is false. The check SHALL fail when `messageTokenThreshold`, `observationTokenThreshold`, or `maxMessageTokenBudget` are non-positive. The check SHALL warn when `maxMessageTokenBudget` is not greater than `messageTokenThreshold`. The check SHALL warn when a custom provider is specified but not found in the providers map.

#### Scenario: OM disabled
- **WHEN** `observationalMemory.enabled` is false
- **THEN** the check returns StatusSkip with message "Observational memory is disabled"

#### Scenario: Invalid thresholds
- **WHEN** `messageTokenThreshold` is 0 or negative
- **THEN** the check returns StatusFail with a message identifying the invalid field

#### Scenario: Budget less than threshold
- **WHEN** `maxMessageTokenBudget` is less than or equal to `messageTokenThreshold`
- **THEN** the check returns StatusWarn indicating the inconsistency

#### Scenario: Unknown provider
- **WHEN** `provider` is set to "custom-llm" but no such provider exists in `providers` map
- **THEN** the check returns StatusWarn indicating the provider was not found

#### Scenario: Valid configuration
- **WHEN** all thresholds are positive, budget exceeds threshold, and provider exists
- **THEN** the check returns StatusPass

### Requirement: Output Scanning diagnostic check
The system SHALL include an OutputScanningCheck in the doctor command that validates interceptor and secret alignment. The check SHALL skip when the interceptor is disabled and no secrets exist. The check SHALL warn when the interceptor is disabled but secrets exist in the database. The check SHALL warn when the interceptor is enabled but PII redaction is disabled. The check SHALL pass when both interceptor and PII redaction are enabled. The check SHALL handle encrypted databases gracefully by returning StatusSkip.

#### Scenario: Interceptor disabled with no secrets
- **WHEN** `security.interceptor.enabled` is false and no secrets are stored
- **THEN** the check returns StatusSkip

#### Scenario: Interceptor disabled with secrets
- **WHEN** `security.interceptor.enabled` is false but secrets exist in the database
- **THEN** the check returns StatusWarn indicating secrets will not be redacted

#### Scenario: Interceptor enabled without PII redaction
- **WHEN** `security.interceptor.enabled` is true but `redactPii` is false
- **THEN** the check returns StatusWarn

#### Scenario: Fully configured
- **WHEN** both interceptor and PII redaction are enabled
- **THEN** the check returns StatusPass

#### Scenario: Encrypted database
- **WHEN** the session database is encrypted and cannot be opened
- **THEN** the check returns StatusSkip with message "Cannot verify (database encrypted)"

### Requirement: Doctor check registration
The system SHALL register ObservationalMemoryCheck and OutputScanningCheck in the AllChecks() function so they are executed by the `lango doctor` command.

#### Scenario: Doctor runs all checks
- **WHEN** user runs `lango doctor`
- **THEN** the output includes results for "Observational Memory" and "Output Scanning" checks
