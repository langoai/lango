## Why

`lango account paymaster approve` still writes table and JSON output directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the smart account CLI writer-routing hardening work.

## What Changes

- Route `lango account paymaster approve` table and JSON output through `cmd.OutOrStdout()`
- Add a small paymaster-approval seam for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the smart account paymaster approval surface consistent with the CLI writer-routing hardening work
- Improves testability without changing approval semantics
