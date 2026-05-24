## Why

`lango p2p peers` still writes empty-state, table, and JSON output directly to process stdout. That makes wrappers and tests depend on global stream interception instead of standard Cobra command capture.

## What Changes

- Route `lango p2p peers` output through `cmd.OutOrStdout()`
- Add command-level regression coverage for empty, table, and JSON paths
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes `p2p peers` consistent with the CLI writer-routing hardening work
- Improves testability without changing payload shape or peer discovery behavior
