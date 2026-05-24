## Why

`lango account deploy` still writes table and JSON output directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the smart account CLI writer-routing hardening work.

## What Changes

- Route `lango account deploy` table and JSON output through `cmd.OutOrStdout()`
- Add a small deploy-result seam for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the smart account deployment surface consistent with the CLI writer-routing hardening work
- Improves testability without changing deployment semantics
