## Why

`lango metrics` still writes table and JSON output directly to process stdout. As the top-level observability snapshot command, it should support standard Cobra output capture for wrappers and tests.

## What Changes

- Route `lango metrics` table and JSON output through `cmd.OutOrStdout()`
- Add command-level regression coverage using an `httptest` gateway
- Update docs and OpenSpec with the output-writer contract

## Impact

- Makes the top-level metrics summary consistent with the CLI writer-routing hardening work
- Improves testability without changing the `/metrics` payload semantics
