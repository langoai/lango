## Why

`lango metrics policy` still writes table and JSON output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango metrics policy` table and JSON output through `cmd.OutOrStdout()`
- Add command-level regression coverage using an `httptest` gateway
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the policy metrics surface consistent with the CLI writer-routing hardening work
- Improves testability without changing the gateway payload shape
