## Why

`app.New()` was a 900+ line sequential initializer while `modules.go` (Phase 4 of the appinit module system) existed as dead code. Additionally, OpenSpec specs for `config-default-walker` and `cli-bootstrap-factory` had drifted from their implementations, and downstream docs had undocumented default values.

## What Changes

- **Spec/Code alignment**: Update `config-default-walker` spec to match actual unexported `setDefaultsFromStruct` implementation; fix `serveCmd()` to use `cliboot.BootResult()` per `cli-bootstrap-factory` spec
- **Module build transition**: Rewrite `app.New()` to use `appinit.Builder.Build()` with 5 modules (foundation, intelligence, automation, network, extension), replacing the monolithic sequential initializer
- **Module parity**: Complete all 5 modules with missing features (workspace, nonce cache, health monitor, team-escrow bridges, lifecycle components) to achieve 1:1 parity with the old `app.New()`
- **Dead code removal**: Delete `registerLifecycleComponents()`, `registerEconomyLifecycle()`, `registerObservabilityLifecycle()` after module transition
- **Docs sync**: Add OutputManager config section, fix Presidio/Session/P2P default value mismatches in `docs/configuration.md`

## Capabilities

### New Capabilities
- `app-module-build`: Module-based application initialization via `appinit.Builder.Build()` replacing monolithic `app.New()`

### Modified Capabilities
- `config-default-walker`: Spec updated to match actual unexported implementation (no exported `WalkDefaults`, maps skipped)
- `cli-bootstrap-factory`: Code fixed to match spec requirement (no direct `bootstrap.Run()` in `cmd/`)

## Impact

- `internal/app/app.go` — `New()` rewritten (~250 lines, down from ~900), 6 helper functions extracted
- `internal/app/modules.go` — All 5 modules updated with lifecycle components, workspace, team-escrow bridges
- `internal/appinit/builder.go` — `BuildResult.CatalogEntries` added
- `internal/appinit/module.go` — `ProvidesBaseTools` key added
- `cmd/lango/main.go` — `serveCmd()` bootstrap call changed
- `openspec/specs/config-default-walker/spec.md` — Rewritten
- `docs/configuration.md` — OutputManager section added, 6 default values corrected
- Dead code removed from `wiring_economy.go`, `wiring_observability.go`
