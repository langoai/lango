## Why

`lango p2p workspace` guidance commands still write text and JSON output directly to process stdout. That makes command output harder to capture in wrappers and tests.

## What Changes

- Route `lango p2p workspace` guidance text and JSON output through `cmd.OutOrStdout()`
- Add regression coverage for the JSON guidance paths
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the workspace guidance surface consistent with the CLI writer-routing hardening work
- Improves testability without changing the guidance wording
