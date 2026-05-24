## Why

`lango p2p firewall` subcommands still write empty-state, table, JSON, and guidance output directly to process stdout. That makes command output harder to capture in wrappers and tests.

## What Changes

- Route `lango p2p firewall list`, `add`, and `remove` output through `cmd.OutOrStdout()`
- Add command-level regression coverage for empty-state, table, JSON, and guidance paths
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the firewall management surface consistent with the CLI writer-routing hardening work
- Improves testability without changing firewall rule semantics
