## Why

`lango account session create`, `list`, and `revoke` still write output directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the smart account CLI writer-routing hardening work.

## What Changes

- Route `lango account session create` table and JSON output through `cmd.OutOrStdout()`
- Route `lango account session list` table, JSON, and empty-state output through `cmd.OutOrStdout()`
- Route `lango account session revoke` success output through `cmd.OutOrStdout()`
- Add small session seams for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the smart account session surface consistent with the CLI writer-routing hardening work
- Improves testability without changing session semantics
