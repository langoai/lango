## Why

`lango p2p pricing` still writes text and JSON output directly to process stdout. The public docs cover the command, but the main P2P CLI spec does not define the pricing command contract.

## What Changes

- Route `lango p2p pricing` output through `cmd.OutOrStdout()`
- Add command-level regression coverage for overview, tool-specific, and JSON output
- Add the missing pricing command contract to the main P2P CLI spec
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes `p2p pricing` consistent with the CLI writer-routing hardening work
- Closes a spec gap between the implemented command and the main P2P CLI capability contract
