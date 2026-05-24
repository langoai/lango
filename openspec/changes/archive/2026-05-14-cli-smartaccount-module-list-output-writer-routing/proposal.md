## Why

`lango account module list` still writes table, JSON, and empty-state output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango account module list` table, JSON, and empty-state output through `cmd.OutOrStdout()`
- Add a small module-list seam for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the smart account module inspection surface consistent with the CLI writer-routing hardening work
- Improves testability without changing module-list semantics
