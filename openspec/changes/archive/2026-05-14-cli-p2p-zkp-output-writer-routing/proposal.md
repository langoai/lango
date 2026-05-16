## Why

`lango p2p zkp status` and `lango p2p zkp circuits` still write text, table, and JSON output directly to process stdout. The public docs are also stale relative to the actual output shape.

## What Changes

- Route `lango p2p zkp status` and `circuits` output through `cmd.OutOrStdout()`
- Add command-level regression coverage for text, table, and JSON paths
- Update docs and OpenSpec with the writer-routing contract and current output shape

## Impact

- Makes the ZKP inspection surface consistent with the CLI writer-routing hardening work
- Improves testability and removes stale public documentation
