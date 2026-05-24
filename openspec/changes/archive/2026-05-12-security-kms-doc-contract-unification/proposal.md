## Why

The public security docs now describe the supported signer-provider set correctly, but KMS setup guidance is still uneven: one README table row was stale, and the CLI/security docs did not consistently say that KMS providers need both the matching build tag and bootstrap-backed storage wiring.

## What Changes

- Unify the remaining README security-provider row with the current signer-provider contract.
- Update README, CLI security docs, and encryption docs to state that KMS providers require the matching build tag and bootstrap-backed storage wiring.
- Sync downstream-docs spec coverage for that KMS setup contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `downstream-docs-sync`: public KMS signer docs now describe the build-tag and bootstrap-wiring prerequisites consistently.

## Impact

- Affected docs: `README.md`, `docs/cli/security.md`, `docs/security/encryption.md`
- Affected specs: `openspec/specs/downstream-docs-sync/spec.md`
