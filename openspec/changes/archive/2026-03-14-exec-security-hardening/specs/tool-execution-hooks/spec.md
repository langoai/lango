## MODIFIED Requirements

### Requirement: SecurityFilterHook blocks dangerous command patterns
The SecurityFilterHook SHALL include a set of default blocked patterns that are always active regardless of user configuration. Default patterns SHALL include catastrophic operations: `rm -rf /`, `mkfs.`, `dd if=/dev/zero`, fork bomb, `> /dev/sda`, `chmod -R 777 /`, `dd if=/dev/random`, `mv / `. User-configured patterns SHALL be merged with defaults, with case-insensitive deduplication. All patterns SHALL be pre-lowercased at construction time to avoid repeated lowercasing in the Pre() hot path.

#### Scenario: Default pattern blocks rm -rf
- **WHEN** agent executes `rm -rf /` via exec tool
- **THEN** SecurityFilterHook blocks it with reason "command matches blocked pattern: rm -rf /"

#### Scenario: User patterns merged with defaults
- **WHEN** SecurityFilterHook is constructed with user pattern "DROP TABLE"
- **THEN** both default patterns and "DROP TABLE" are active

#### Scenario: Duplicate patterns deduplicated
- **WHEN** user configures "rm -rf /" which is already a default
- **THEN** the pattern appears only once in the merged list

### Requirement: SecurityFilterHook always registered
The SecurityFilterHook SHALL be registered unconditionally in the tool hook pipeline, not gated by `cfg.Hooks.Enabled` or `cfg.Hooks.SecurityFilter`. Other hooks (AccessControl, EventPublishing) remain config-gated.

#### Scenario: Security hook active without config
- **WHEN** hooks.enabled is false and hooks.securityFilter is false
- **THEN** SecurityFilterHook is still active with default patterns
