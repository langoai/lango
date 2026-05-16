## Why

`lango p2p session` subcommands still write table output, JSON output, and revoke confirmations directly to process stdout. The public docs cover the commands, but the main P2P CLI spec does not define the session command group contract.

## What Changes

- Route `lango p2p session list`, `revoke`, and `revoke-all` output through `cmd.OutOrStdout()`
- Add command-level regression coverage for empty-state, table, JSON, and confirmation paths
- Add the missing session command contract to the main P2P CLI spec
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the session command group consistent with the CLI writer-routing hardening work
- Closes a spec gap between the implemented command group and the main P2P CLI capability contract
