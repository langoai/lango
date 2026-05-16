## Why

`lango account info` still writes table and JSON output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango account info` table and JSON output through `cmd.OutOrStdout()`
- Add a small info-result seam for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the smart account info surface consistent with the CLI writer-routing hardening work
- Improves testability without changing smart account info semantics
