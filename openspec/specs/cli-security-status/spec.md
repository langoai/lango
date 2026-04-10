## Purpose

Capability spec for cli-security-status. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Security status command

The system SHALL provide a `lango security status` command that displays the current security configuration and state. The command SHALL show signer provider, encryption key count, stored secret count, interceptor status, PII redaction status, approval policy, DB encryption state, and envelope information (version, KEK slot count/types, recovery setup, pending flags). The command SHALL support `--json` for JSON output. The command default behavior SHALL be passphrase-free: it SHALL NOT trigger an interactive passphrase prompt. When DB access requires a passphrase that cannot be obtained non-interactively (via keyring or keyfile), the command SHALL gracefully degrade DB-dependent fields (e.g., encryption key count = 0, signer provider = "unavailable") without failing.

#### Scenario: Display security status with envelope fields

- **WHEN** user runs `lango security status` with an envelope-based installation
- **THEN** the command SHALL display the envelope version, number of KEK slots, slot types (passphrase, mnemonic), recovery setup status, and any pending flags (`PendingMigration`, `PendingRekey`)

#### Scenario: Display security status with approval policy

- **WHEN** user runs `lango security status`
- **THEN** the command SHALL display "Approval Policy: <policy>" where policy is the `ApprovalPolicy` string value (defaulting to "dangerous" if empty)

#### Scenario: JSON output with envelope and approval policy

- **WHEN** user runs `lango security status --json`
- **THEN** the JSON output SHALL include envelope fields (`envelope_version`, `kek_slots`, `recovery_setup`, `pending_migration`, `pending_rekey`) and `"approval_policy": "<policy>"`

#### Scenario: Database unavailable (non-interactive)

- **WHEN** the session database cannot be opened because no passphrase is available non-interactively
- **THEN** the command displays all envelope fields and configuration fields
- **AND** DB-dependent fields show zero counts or "unavailable"
- **AND** the command exits with code 0 without failing

#### Scenario: Passphrase-free default behavior

- **WHEN** user runs `lango security status` in any environment
- **THEN** the command SHALL NOT trigger an interactive passphrase prompt
- **AND** it SHALL use `passphrase.AcquireNonInteractive()` (keyring/keyfile only)
- **AND** if neither source provides a passphrase, it SHALL proceed with DB fields unavailable

#### Scenario: Database unavailable (legacy behavior preserved)

- **WHEN** the session database cannot be opened (any reason)
- **THEN** the command displays status with zero counts for keys and secrets, without failing

### Requirement: Non-interactive mini-bootstrap for status

The system SHALL provide a `readDBStatusNonInteractive` helper that runs a minimal bootstrap (envelope load → non-interactive passphrase → MK unwrap → read-only DB open → read counts → close) without triggering interactive prompts or schema migration. The helper SHALL handle both envelope-based and legacy installations.

#### Scenario: Envelope-based non-interactive read

- **WHEN** `readDBStatusNonInteractive` is called with an envelope present and a keyring-stored passphrase
- **THEN** the helper unwraps the MK, derives the DB key via `DeriveDBKeyHex(mk)`, opens the DB read-only with `PRAGMA key = "x'<hex>'"`, reads key and secret counts, and closes the DB

#### Scenario: Legacy non-interactive read

- **WHEN** `readDBStatusNonInteractive` is called with no envelope and a keyfile-stored passphrase
- **THEN** the helper uses the passphrase directly as the DB key, opens the DB read-only, reads counts, and closes

#### Scenario: No passphrase available

- **WHEN** `readDBStatusNonInteractive` is called and `AcquireNonInteractive` returns an error
- **THEN** the helper returns a zero-valued `dbStatusResult` (all counts 0)
- **AND** no DB open attempt is made

### Requirement: Read-only database open for status

The system SHALL provide an `OpenDatabaseReadOnly` function used by status commands. This function SHALL open the SQLite database in read-only mode (`file:path?mode=ro`), SHALL NOT invoke ent schema migration, and SHALL NOT create any tables or indexes. The contract: read-only, no migration, no prompt, failure returns error (caller degrades gracefully).

#### Scenario: Read-only open succeeds

- **WHEN** `OpenDatabaseReadOnly(dbPath, dbKey, rawKey)` is called with a valid DB and key
- **THEN** the function opens the DB with `mode=ro`
- **AND** does NOT call `Schema.Create`
- **AND** returns an `*ent.Client` backed by a read-only connection

#### Scenario: Read-only open rejects writes

- **WHEN** a write operation is attempted on the read-only client
- **THEN** SQLite returns a "read-only database" error
