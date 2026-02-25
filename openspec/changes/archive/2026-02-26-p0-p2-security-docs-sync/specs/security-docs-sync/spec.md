## ADDED Requirements

### Requirement: CLI security docs include OS Keyring commands
The `docs/cli/security.md` file SHALL document `lango security keyring store`, `keyring clear` (with `--force`), and `keyring status` (with `--json`) commands with output examples matching the actual CLI implementation.

#### Scenario: Keyring commands documented
- **WHEN** a user reads `docs/cli/security.md`
- **THEN** they find complete documentation for `keyring store`, `keyring clear`, and `keyring status` with flags, examples, and JSON output fields

### Requirement: CLI security docs include DB encryption commands
The `docs/cli/security.md` file SHALL document `lango security db-migrate` and `lango security db-decrypt` commands with `--force` flag and output examples.

#### Scenario: DB encryption commands documented
- **WHEN** a user reads `docs/cli/security.md`
- **THEN** they find complete documentation for `db-migrate` and `db-decrypt` with flags and examples

### Requirement: CLI security docs include KMS commands
The `docs/cli/security.md` file SHALL document `lango security kms status` (with `--json`), `kms test`, and `kms keys` (with `--json`) commands with output examples.

#### Scenario: KMS commands documented
- **WHEN** a user reads `docs/cli/security.md`
- **THEN** they find complete documentation for `kms status`, `kms test`, and `kms keys` with JSON output fields

### Requirement: CLI security status output includes new fields
The `docs/cli/security.md` status example SHALL include `DB Encryption`, `KMS Provider`, `KMS Key ID`, and `KMS Fallback` fields matching `status.go` output.

#### Scenario: Updated status output documented
- **WHEN** a user reads the `security status` example
- **THEN** they see all fields including `db_encryption`, `kms_provider`, `kms_key_id`, `kms_fallback` in the JSON fields table

### Requirement: CLI P2P docs include session management commands
The `docs/cli/p2p.md` file SHALL document `lango p2p session list` (with `--json`), `session revoke` (with `--peer-did`), and `session revoke-all` commands.

#### Scenario: Session commands documented
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** they find complete documentation for session list, revoke, and revoke-all

### Requirement: CLI P2P docs include sandbox commands
The `docs/cli/p2p.md` file SHALL document `lango p2p sandbox status`, `sandbox test`, and `sandbox cleanup` commands with output examples.

#### Scenario: Sandbox commands documented
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** they find complete documentation for sandbox status, test, and cleanup

### Requirement: Feature docs cover signed handshake protocol
The `docs/features/p2p-network.md` SHALL document the signed challenge protocol (v1.0/v1.1), ECDSA signature, timestamp validation, and nonce replay protection.

#### Scenario: Signed handshake documented
- **WHEN** a user reads the Handshake section
- **THEN** they understand protocol versioning, signed challenges, and `requireSignedChallenge` config

### Requirement: Feature docs cover session management
The `docs/features/p2p-network.md` SHALL include a Session Management section with invalidation reasons and SecurityEventHandler.

#### Scenario: Session management documented
- **WHEN** a user reads P2P feature docs
- **THEN** they find session invalidation reasons, auto-revocation triggers, and CLI commands

### Requirement: Feature docs cover tool sandbox
The `docs/features/p2p-network.md` SHALL include a Tool Execution Sandbox section with isolation modes, runtime probe chain, and container pool.

#### Scenario: Tool sandbox documented
- **WHEN** a user reads P2P feature docs
- **THEN** they find subprocess/container modes, runtime probe chain, and configuration

### Requirement: Feature docs cover credential revocation
The `docs/features/p2p-network.md` SHALL include a Credential Revocation section with RevokeDID, IsRevoked, and maxCredentialAge.

#### Scenario: Credential revocation documented
- **WHEN** a user reads P2P feature docs
- **THEN** they find revocation mechanisms and credential validation checks

### Requirement: Security index includes new layers
The `docs/security/index.md` SHALL list OS Keyring, Database Encryption, Cloud KMS/HSM, P2P Session Management, P2P Tool Sandbox, and P2P Auth Hardening in the Security Layers table.

#### Scenario: Security layers table updated
- **WHEN** a user reads the security index
- **THEN** they see all 10 security layers including the 6 new ones

### Requirement: Encryption docs cover Cloud KMS
The `docs/security/encryption.md` SHALL include a Cloud KMS Mode section with all 4 backends, build tags, CompositeCryptoProvider, and configuration examples.

#### Scenario: Cloud KMS documented
- **WHEN** a user reads encryption docs
- **THEN** they find all 4 KMS backends with configuration examples

### Requirement: README config table includes new keys
The `README.md` configuration table SHALL include all P2P security, tool isolation, ZKP, keyring, DB encryption, and KMS config keys matching `mapstructure` tags.

#### Scenario: Config table complete
- **WHEN** a user reads the README config table
- **THEN** they find 27+ new config rows covering all P0-P2 security features

### Requirement: Agent prompts include P2P security awareness
The `prompts/AGENTS.md` and `prompts/TOOL_USAGE.md` SHALL include references to signed challenges, session management, sandbox, KMS, and credential revocation.

#### Scenario: Agent prompts updated
- **WHEN** the LLM agent loads prompts
- **THEN** it has awareness of all P0-P2 security features

### Requirement: Security roadmap P0/P1 items marked complete
The `openspec/security-roadmap.md` SHALL have `✅ COMPLETED` markers on all P0 and P1 section headers.

#### Scenario: Roadmap completion markers
- **WHEN** a user reads the security roadmap
- **THEN** all P0 (P0-1, P0-2, P0-3) and P1 (P1-4, P1-5, P1-6) items show completion markers
