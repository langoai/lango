## Why

P0-P2 security hardening implementation is complete in code, but documentation, agent prompts, and UI-facing references have not been updated. Users and agents cannot discover the new security features (OS Keyring, DB Encryption, Cloud KMS, Session Management, Tool Sandbox, Signed Challenges, Credential Revocation) through docs or prompts, creating a gap between implementation and discoverability.

## What Changes

- Update `docs/cli/security.md` with new CLI commands: `keyring store/clear/status`, `db-migrate`, `db-decrypt`, `kms status/test/keys`, and updated `status` output
- Update `docs/cli/p2p.md` with new CLI commands: `session list/revoke/revoke-all`, `sandbox status/test/cleanup`
- Update `docs/security/encryption.md` with Cloud KMS Mode, OS Keyring Integration, Database Encryption sections
- Update `docs/security/index.md` with 6 new security layers and Cloud KMS encryption mode
- Update `docs/features/p2p-network.md` with signed challenges, session management, tool sandbox, ZK circuit updates, credential revocation
- Update `README.md` with 27+ new config rows, feature bullets, CLI examples, and security subsections
- Update `prompts/AGENTS.md` with expanded P2P Network description
- Update `prompts/TOOL_USAGE.md` with session management, sandbox, signed challenge, KMS, and credential revocation guidance
- Update `openspec/security-roadmap.md` with P0/P1 completion markers

## Capabilities

### New Capabilities
- `security-docs-sync`: Documentation synchronization for all P0-P2 security features across CLI docs, feature docs, README, agent prompts, and security roadmap

### Modified Capabilities

## Impact

- 9 documentation/prompt files modified (~545 lines added/modified)
- No code changes — documentation-only change
- All CLI output examples verified against actual source code (status.go, keyring.go, kms.go, db_migrate.go, session.go, sandbox.go)
- All config keys verified against mapstructure tags in internal/config/types.go
