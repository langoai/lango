## Why

`lango security migrate-passphrase` still writes its guidance, progress, and success output directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the CLI writer-routing hardening work.

## What Changes

- Route `lango security migrate-passphrase` non-error status output through `cmd.OutOrStdout()`
- Pass the Cobra writer down into the migration helper so progress messages stay on the command stream
- Add a small execution seam for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the legacy passphrase migration surface consistent with the CLI writer-routing hardening work
- Improves testability without changing migration semantics
