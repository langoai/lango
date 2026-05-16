## Why

`lango p2p team` guidance commands still write text and JSON output directly to process stdout. That makes command output harder to capture in wrappers and tests.

## What Changes

- Route `lango p2p team list`, `status`, and `disband` output through `cmd.OutOrStdout()`
- Add regression coverage for JSON guidance output
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the team guidance surface consistent with the CLI writer-routing hardening work
- Improves testability without changing the guidance content
