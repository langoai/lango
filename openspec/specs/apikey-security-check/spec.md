## Purpose

Capability spec for apikey-security-check. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: API Key Security Diagnostic Check
The system SHALL provide a diagnostic check in `lango doctor` that validates API key storage. Keys stored in encrypted profiles are considered safe. Keys not using `${ENV_VAR}` references are flagged as "inline" with guidance that encrypted profiles are safe and `${ENV_VAR}` references are an alternative for portability.

#### Scenario: All keys use environment variable references
- **WHEN** all configured providers use `${ENV_VAR}` pattern for `apiKey`
- **THEN** the check SHALL report `StatusPass` with message "All API keys secured"

#### Scenario: Inline API key detected
- **WHEN** one or more providers have an `apiKey` that does not match `${...}` pattern
- **THEN** the check SHALL report `StatusWarn`
- **AND** the message SHALL read "Inline API keys for: <provider-ids>"
- **AND** details SHALL state that keys in encrypted profiles are safe and suggest `${ENV_VAR}` references for portability

#### Scenario: No providers configured
- **WHEN** the `providers` map is empty
- **THEN** the check SHALL report `StatusSkip`

#### Scenario: No API keys set
- **WHEN** providers exist but all have empty `apiKey` values
- **THEN** the check SHALL report `StatusSkip`

#### Scenario: Configuration not loaded
- **WHEN** config is nil
- **THEN** the check SHALL report `StatusSkip`
