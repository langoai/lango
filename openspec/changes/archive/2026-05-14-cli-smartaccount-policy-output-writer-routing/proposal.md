## Why

`lango account policy show` and `lango account policy set` still write output directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the smart account CLI writer-routing hardening work.

## What Changes

- Route `lango account policy show` table and JSON output through `cmd.OutOrStdout()`
- Route `lango account policy set` success output through `cmd.OutOrStdout()`
- Add small policy seams for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the smart account policy surface consistent with the CLI writer-routing hardening work
- Improves testability without changing policy semantics
