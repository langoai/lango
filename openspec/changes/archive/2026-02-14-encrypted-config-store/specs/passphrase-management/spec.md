## MODIFIED Requirements

### Requirement: Passphrase source resolution
The system SHALL resolve passphrases using the priority chain: keyfile (`~/.lango/keyfile`) → interactive terminal prompt → stdin pipe. The system SHALL NOT read passphrases from the `LANGO_PASSPHRASE` environment variable or the `security.passphrase` config field.

#### Scenario: Passphrase acquisition in CLI security commands
- **WHEN** `initLocalCrypto` is called in CLI security commands
- **THEN** the passphrase is acquired via `passphrase.Acquire()` (not env var or config)

#### Scenario: Non-interactive environment without keyfile
- **WHEN** stdin is not a terminal and no keyfile exists
- **THEN** the system attempts to read from stdin pipe; if empty, returns an error

## REMOVED Requirements

### Requirement: LANGO_PASSPHRASE environment variable support
**Reason**: Environment variables are visible to child processes and can be leaked through the exec tool. Replaced by keyfile-based non-interactive authentication.
**Migration**: Create `~/.lango/keyfile` with 0600 permissions containing the passphrase. In CI/Docker: `echo "your-passphrase" > ~/.lango/keyfile && chmod 600 ~/.lango/keyfile`
