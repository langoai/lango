## Why

The config/bootstrap path has 4 regressions on the dev branch: normalization+validation is scattered across multiple call sites, leading to inconsistent behavior. Bootstrap profile loading skips PostLoad, Store.Save() only validates but doesn't normalize, config set uses double bootstrap (causing keyfile shred failures), and the collaborator preset fails validation because it enables payment without an RPC URL. The core principle is: "normalize+validate in one place, CLI reuses that result once."

## What Changes

- Add `config.PostLoad()` one-stop function that chains migration, env substitution, path normalization, path validation, and config validation
- Refactor `config.Load()` to use `PostLoad` instead of 5 separate calls
- Remove early returns in `phaseLoadProfile()` so all branches go through a single PostLoad call at the end
- Add `PostLoad()` call at the start of `Store.Save()` so persisted configs are always canonical
- Change `NewSetCmd` cfgLoader signature to return a cleanup function, eliminating double bootstrap and DB client leaks
- Add `payment.network.rpcUrl` to the collaborator preset to match the default ChainID (Base Sepolia 84532)

## Capabilities

### New Capabilities

### Modified Capabilities
- `config-system`: PostLoad() added as the single normalization+validation entry point; Load() delegates to it
- `bootstrap-lifecycle`: phaseLoadProfile() applies PostLoad after all branches instead of early returns
- `encrypted-config-profiles`: Store.Save() runs PostLoad before persisting to ensure canonical form
- `config-cli-commands`: config set uses cleanup-function pattern to prevent double bootstrap and DB leaks
- `config-presets`: collaborator preset includes RPC URL for Base Sepolia

## Impact

- `internal/config/loader.go` — new exported `PostLoad()` function, `Load()` simplified
- `internal/bootstrap/phases.go` — `phaseLoadProfile()` restructured, new import `config`
- `internal/configstore/store.go` — `Save()` calls `PostLoad()` before marshal
- `internal/cli/configcmd/getset.go` — `NewSetCmd` cfgLoader signature changed (breaking for callers)
- `cmd/lango/main.go` — config set wiring updated to single bootstrap + cleanup closure
- `internal/config/presets.go` — collaborator preset adds RPCURL field
