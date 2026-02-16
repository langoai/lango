## MODIFIED Requirements

### Requirement: Security status command
The system SHALL provide a `lango security status` command that displays the current security configuration and state. The command SHALL show signer provider, encryption key count, stored secret count, interceptor status, PII redaction status, and approval policy. The command SHALL support `--json` for JSON output. The command SHALL NOT require a passphrase (it reads counts only, no decryption).

#### Scenario: Display security status with approval policy
- **WHEN** user runs `lango security status`
- **THEN** the command SHALL display "Approval Policy: <policy>" where policy is the `ApprovalPolicy` string value (defaulting to "dangerous" if empty)

#### Scenario: JSON output with approval policy
- **WHEN** user runs `lango security status --json`
- **THEN** the JSON output SHALL include `"approval_policy": "<policy>"` instead of the legacy `"approval_required"` field

#### Scenario: Database unavailable
- **WHEN** the session database cannot be opened
- **THEN** the command displays status with zero counts for keys and secrets, without failing
