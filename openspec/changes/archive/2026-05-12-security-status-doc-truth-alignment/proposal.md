## Why

The `lango security status` command already exposes richer semantics than the docs currently describe. In particular, `signer_provider`, `db_encryption`, and `kms_fallback` are documented too narrowly, which makes the CLI reference lag behind the actual runtime contract.

## What Changes

- Update the `lango security status` JSON field docs to match current output semantics.
- Sync `downstream-docs-sync` so the security status reference stays aligned with the actual command contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `downstream-docs-sync`: security status docs now describe the real field semantics for provider, DB protection state, and KMS fallback.

## Impact

- Affected docs: `docs/cli/security.md`
- Affected specs: `openspec/specs/downstream-docs-sync/spec.md`
