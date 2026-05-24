## Why

`lango p2p connect` and `lango p2p disconnect` still print success confirmations directly to process stdout. That prevents wrappers and command-level tests from capturing the output through normal Cobra streams.

## What Changes

- Route `lango p2p connect` and `lango p2p disconnect` success confirmations through `cmd.OutOrStdout()`
- Add command-level regression coverage for both success paths
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes these control commands consistent with the CLI writer-routing hardening work
- Improves testability without changing connect/disconnect behavior
