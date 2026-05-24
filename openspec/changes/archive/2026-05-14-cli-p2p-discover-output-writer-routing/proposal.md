## Why

`lango p2p discover` still writes empty-state, table, and JSON output directly to process stdout. That prevents wrappers and tests from capturing command output through the standard Cobra streams.

## What Changes

- Route `lango p2p discover` output through `cmd.OutOrStdout()`
- Add command-level regression coverage for empty, table, and JSON output
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes `p2p discover` consistent with the rest of the CLI writer-routing hardening work
- Improves testability without changing discovery payloads or filtering semantics
