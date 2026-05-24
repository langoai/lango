## Summary

Convert `session.EntStore.MigrateSecrets` panic recovery into an ordinary returned error while preserving transaction rollback behavior.

## Motivation

The deprecated `lango security migrate-passphrase` path still calls `MigrateSecrets`. A panic from the re-encryption callback or transaction internals currently rolls back and then re-panics, crashing the CLI instead of returning the existing `migration failed: ...` error shape.

## Scope

- Preserve rollback-on-panic behavior inside `MigrateSecrets`.
- Convert recovered panics into returned errors after rollback.
- Add a regression test for a panicking re-encryption callback.
- Sync and archive the OpenSpec change after verification.

## Non-Goals

- No changes to `change-passphrase`.
- No changes to key derivation, encryption format, or stored secret schema.
- No removal of the deprecated `migrate-passphrase` command.
