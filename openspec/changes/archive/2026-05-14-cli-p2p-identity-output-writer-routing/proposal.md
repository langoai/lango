## Why

`lango p2p identity` still writes text and JSON output directly to process stdout. That makes command output difficult to capture in wrappers and tests, and the public docs are stale about whether the CLI surfaces the DID.

## What Changes

- Route `lango p2p identity` text and JSON output through `cmd.OutOrStdout()`
- Add command-level regression coverage for text and JSON writer capture
- Update docs and OpenSpec to reflect the active DID output and command-writer contract

## Impact

- Makes `p2p identity` consistent with the writer-routing hardening work across the CLI
- Fixes stale public documentation about the identity command
- Improves testability without changing the identity payload shape
