## Why

`lango p2p provenance push` and `fetch` still print success confirmations directly to process stdout. The public docs mention the commands, but the main P2P CLI spec does not define the provenance command group contract.

## What Changes

- Route provenance push/fetch success confirmations through `cmd.OutOrStdout()`
- Add command-level regression coverage for both success paths
- Add the missing provenance command contract to the main P2P CLI spec
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes provenance exchange commands consistent with the CLI writer-routing hardening work
- Closes a spec gap between the implemented command group and the main P2P CLI capability contract
