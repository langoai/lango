## Context

`internal/session/ent_store.go` uses a transaction in `MigrateSecrets` to re-encrypt secrets and update passphrase metadata. Its deferred cleanup rolls back on panic, then rethrows the panic.

## Design

Keep the named return value on `MigrateSecrets` and update the deferred transaction cleanup so recovered panics are converted into an error after rollback. The error should identify the migration action and include the recovered panic value. Existing non-panic errors keep the current rollback behavior and return path.

This keeps the CLI-facing error pipeline intact: `migrateSecrets` continues wrapping store failures as `migration failed: ...`, but the process no longer crashes on callback panics.

## Testing

Add a session package regression that creates an Ent-backed store with a stored secret, calls `MigrateSecrets` with a panicking callback, and asserts the call returns an error instead of panicking.

## Risks

The main risk is accidentally committing partial secret updates after a panic. The implementation must rollback before assigning the recovered-panic error, and the test should exercise the transaction path by ensuring at least one secret exists.
