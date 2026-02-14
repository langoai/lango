## ADDED Requirements

### Requirement: Security status command
The system SHALL provide a `lango security status` command that displays the current security configuration and state. The command SHALL show signer provider, encryption key count, stored secret count, interceptor status, PII redaction status, and approval required status. The command SHALL support `--json` for JSON output. The command SHALL NOT require a passphrase (it reads counts only, no decryption).

#### Scenario: Display security status
- **WHEN** user runs `lango security status`
- **THEN** the command displays signer provider name, key count, secret count, and interceptor configuration values

#### Scenario: JSON output
- **WHEN** user runs `lango security status --json`
- **THEN** the command outputs a JSON object with all status fields

#### Scenario: Database unavailable
- **WHEN** the session database cannot be opened
- **THEN** the command displays status with zero counts for keys and secrets, without failing
