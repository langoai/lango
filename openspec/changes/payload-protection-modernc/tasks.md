## 1. Driver And Legacy Policy

- [x] 1.1 Introduce a SQLite driver adapter and switch the primary runtime to modernc.
- [x] 1.2 Remove SQLCipher runtime paths and DB encrypt/decrypt workflows.
- [x] 1.3 Detect non-SQLite legacy DB headers and fail fast with remediation messaging.

## 2. Payload Protection

- [x] 2.1 Add encrypted payload fields for user-content-heavy entities.
- [x] 2.2 Implement broker-managed AEAD encrypt/decrypt with `key_version=1`.
- [ ] 2.3 Define redacted search/recalI projection generation and atomic commit rules.

## 3. Verification

- [ ] 3.1 Add round-trip, tamper, and leakage regression tests.
- [ ] 3.2 Verify modernc FTS5 probe/build behavior.
- [ ] 3.3 Update security/docs/status surfaces to reflect brokered payload protection.
