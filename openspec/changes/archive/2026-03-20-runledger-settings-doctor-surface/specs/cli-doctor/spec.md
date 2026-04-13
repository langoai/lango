## MODIFIED Requirements

### Requirement: Doctor Command Entry Point
The system SHALL include RunLedger diagnostics in the `lango doctor` command output and help text.

#### Scenario: Doctor help text mentions RunLedger
- **WHEN** user runs `lango doctor --help`
- **THEN** the long description SHALL include RunLedger configuration diagnostics among the check families

### Requirement: Doctor check registration
The system SHALL register `RunLedgerCheck` in `checks.AllChecks()` so it is executed by `lango doctor`.

#### Scenario: RunLedger check present in doctor execution
- **WHEN** user runs `lango doctor`
- **THEN** the result set SHALL include a `RunLedger` check entry

## ADDED Requirements

### Requirement: RunLedger diagnostic check
The doctor command SHALL include a `RunLedgerCheck` that validates RunLedger-specific configuration invariants.

#### Scenario: RunLedger disabled
- **WHEN** `runLedger.enabled` is false
- **THEN** the check SHALL return `StatusSkip`

#### Scenario: Invalid stale TTL
- **WHEN** `runLedger.enabled` is true
- **AND** `runLedger.staleTtl <= 0`
- **THEN** the check SHALL return `StatusFail`

#### Scenario: Invalid validator timeout
- **WHEN** `runLedger.enabled` is true
- **AND** `runLedger.validatorTimeout <= 0`
- **THEN** the check SHALL return `StatusFail`

#### Scenario: Invalid history or retry values
- **WHEN** `runLedger.maxRunHistory < 0` or `runLedger.plannerMaxRetries < 0`
- **THEN** the check SHALL return `StatusFail`

#### Scenario: Authoritative read without write-through
- **WHEN** `runLedger.authoritativeRead` is true
- **AND** `runLedger.writeThrough` is false
- **THEN** the check SHALL return `StatusFail`

#### Scenario: Valid RunLedger configuration
- **WHEN** RunLedger is enabled and all invariants are satisfied
- **THEN** the check SHALL return `StatusPass`
