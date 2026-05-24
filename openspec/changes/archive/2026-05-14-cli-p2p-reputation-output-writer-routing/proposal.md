## Why

`lango p2p reputation` still writes missing-record, text, and JSON output directly to process stdout. The public P2P docs mention the command, but the main CLI P2P spec is missing a concrete requirement for it.

## What Changes

- Route `lango p2p reputation` output through `cmd.OutOrStdout()`
- Add command-level regression coverage for missing-record, text, and JSON output
- Add the missing reputation command contract to the main P2P CLI spec
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes `p2p reputation` consistent with the CLI writer-routing hardening work
- Closes a spec gap between the implemented command and the main P2P CLI capability contract
