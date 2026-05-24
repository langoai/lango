## Why

`lango security change-passphrase` still writes its success confirmation directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the CLI writer-routing hardening work.

## What Changes

- Route `lango security change-passphrase` success output through `cmd.OutOrStdout()`
- Add a small execution seam for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the passphrase change surface consistent with the CLI writer-routing hardening work
- Improves testability without changing passphrase rotation semantics
