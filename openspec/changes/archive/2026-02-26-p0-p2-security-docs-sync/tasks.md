## 1. CLI Documentation

- [x] 1.1 Update `docs/cli/security.md` — add DB Encryption and KMS fields to security status output example
- [x] 1.2 Update `docs/cli/security.md` — add JSON fields table entries for `db_encryption`, `kms_provider`, `kms_key_id`, `kms_fallback`
- [x] 1.3 Update `docs/cli/security.md` — add OS Keyring section (store, clear, status commands)
- [x] 1.4 Update `docs/cli/security.md` — add Database Encryption section (db-migrate, db-decrypt commands)
- [x] 1.5 Update `docs/cli/security.md` — add Cloud KMS / HSM section (kms status, test, keys commands)
- [x] 1.6 Update `docs/cli/p2p.md` — add Session Management section (session list, revoke, revoke-all commands)
- [x] 1.7 Update `docs/cli/p2p.md` — add Tool Execution Sandbox section (sandbox status, test, cleanup commands)

## 2. Feature Documentation

- [x] 2.1 Update `docs/features/p2p-network.md` — expand Handshake section with signed challenge protocol and protocol versioning
- [x] 2.2 Update `docs/features/p2p-network.md` — add Session Management section with invalidation reasons and SecurityEventHandler
- [x] 2.3 Update `docs/features/p2p-network.md` — add Tool Execution Sandbox section with isolation modes and container runtime
- [x] 2.4 Update `docs/features/p2p-network.md` — expand ZK Circuits section with attestation freshness and SRS configuration
- [x] 2.5 Update `docs/features/p2p-network.md` — add Credential Revocation section
- [x] 2.6 Update `docs/features/p2p-network.md` — update Configuration JSON with new fields
- [x] 2.7 Update `docs/features/p2p-network.md` — update CLI Commands list

## 3. Security Documentation

- [x] 3.1 Update `docs/security/encryption.md` — add Cloud KMS Mode section with 4 backends
- [x] 3.2 Update `docs/security/encryption.md` — add OS Keyring Integration section
- [x] 3.3 Update `docs/security/encryption.md` — add Database Encryption section
- [x] 3.4 Update `docs/security/encryption.md` — update Configuration Reference JSON
- [x] 3.5 Update `docs/security/index.md` — add 6 new rows to Security Layers table
- [x] 3.6 Update `docs/security/index.md` — add Cloud KMS to Encryption Modes
- [x] 3.7 Update `docs/security/index.md` — update Quick Links

## 4. README

- [x] 4.1 Update `README.md` — add 27+ config rows to P2P Network and Security config tables
- [x] 4.2 Update `README.md` — mark `p2p.keyDir` as deprecated
- [x] 4.3 Update `README.md` — add P2P feature bullets (Signed Challenges, Session Management, etc.)
- [x] 4.4 Update `README.md` — add P2P CLI usage examples (session, sandbox commands)
- [x] 4.5 Update `README.md` — add Security subsections (OS Keyring, DB Encryption, Cloud KMS, P2P Hardening)

## 5. Agent Prompts

- [x] 5.1 Update `prompts/AGENTS.md` — expand P2P Network description with security features
- [x] 5.2 Update `prompts/TOOL_USAGE.md` — add 5 P2P security guidance bullets

## 6. Security Roadmap

- [x] 6.1 Update `openspec/security-roadmap.md` — add ✅ COMPLETED to P0-1, P0-2, P0-3
- [x] 6.2 Update `openspec/security-roadmap.md` — add ✅ COMPLETED to P1-4, P1-5, P1-6

## 7. Verification

- [x] 7.1 Run `go build ./...` to confirm no code breakage
