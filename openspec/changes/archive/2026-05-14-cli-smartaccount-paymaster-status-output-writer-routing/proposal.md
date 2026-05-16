## Why

`lango account paymaster status` still writes table and JSON output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango account paymaster status` table and JSON output through `cmd.OutOrStdout()`
- Add a small paymaster-status seam for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the smart account paymaster inspection surface consistent with the CLI writer-routing hardening work
- Improves testability without changing paymaster status semantics
