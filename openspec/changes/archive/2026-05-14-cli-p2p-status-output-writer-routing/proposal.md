## Why

`lango p2p status` still writes text and JSON output directly to process stdout. That makes wrappers and tests rely on global stream interception instead of normal Cobra command capture.

## What Changes

- Route `lango p2p status` text and JSON output through `cmd.OutOrStdout()`
- Add command-level regression coverage for writer capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes `p2p status` consistent with the rest of the CLI hardening work
- Improves testability without changing the status payload shape
