## Why

Once SQLite ownership is brokered, the remaining security and platform complexity comes from SQLCipher/CGO/runtime PRAGMA paths. We need a broker-compatible payload protection model that works with a pure SQLite FTS5 stack and does not depend on page-level DB encryption.

## What Changes

- Switch the primary SQLite driver to `modernc.org/sqlite`.
- Remove SQLCipher runtime paths and treat encrypted legacy DB headers as unsupported.
- Add broker-managed payload encryption for user-content fields while keeping redacted search projections plaintext.
- Make ciphertext/projection/FTS updates atomic within broker transactions.

## Capabilities

### New Capabilities
- `payload-protection`: broker-managed AEAD encryption of sensitive user content with redacted search projections.

### Modified Capabilities
- `db-encryption`: replaced by brokered payload protection and legacy fail-fast behavior.
- `master-key-envelope`: continues to supply master-key material, but no longer derives SQLCipher page keys.
- `cli-security-status`: reports broker/payload protection state instead of DB encryption state.
- `encrypted-config-profiles`: profile storage remains encrypted by app crypto, not SQLCipher page encryption.

## Impact

- Affected code: SQLite driver bootstrap, security surface, recall/projection generation, entity schemas carrying encrypted payload fields, docs/build tooling.
- Breaking behavior: SQLCipher-encrypted DB files are no longer readable by the new runtime.
