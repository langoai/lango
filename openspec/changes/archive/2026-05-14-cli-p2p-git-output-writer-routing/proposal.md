## Why

`lango p2p git` guidance commands still write text guidance and `log --json` output directly to process stdout. That makes command output harder to capture in wrappers and tests.

## What Changes

- Route `lango p2p git` guidance output through `cmd.OutOrStdout()`
- Route `lango p2p git log --json` through `cmd.OutOrStdout()`
- Add regression coverage for JSON capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the git guidance surface consistent with the CLI writer-routing hardening work
- Improves testability without changing the guidance wording
