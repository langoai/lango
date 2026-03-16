## 1. Spec/Code Alignment (Fix 1)

- [x] 1.1 Update `openspec/specs/config-default-walker/spec.md` — change `WalkDefaults` to `setDefaultsFromStruct`, remove map emission, add map skip requirement
- [x] 1.2 Fix `cmd/lango/main.go` `serveCmd()` — replace `bootstrap.Run()` with `cliboot.BootResult()`
- [x] 1.3 Verify build: `go build ./cmd/lango/...`

## 2. AppInit Extensions (Fix 2 — Steps 2.1-2.2)

- [x] 2.1 Add `CatalogEntries []CatalogEntry` to `BuildResult` in `appinit/builder.go`; collect in `Build()` loop
- [x] 2.2 Add `ProvidesBaseTools Provides = "base_tools"` to `appinit/module.go`
- [x] 2.3 Verify build: `go build ./internal/appinit/...`

## 3. Module Parity Restoration (Fix 2 — Step 2.3)

- [x] 3.1 Foundation module: add `ProvidesBaseTools` to Values map
- [x] 3.2 Intelligence module: resolve base tools via `ProvidesBaseTools`, add `Observer` field, add lifecycle ComponentEntry for 5 buffers
- [x] 3.3 Automation module: add lifecycle ComponentEntry for cron-scheduler, background-manager, workflow-engine
- [x] 3.4 Network module: add workspace init + tools + lifecycle, nonce cache, health monitor, team-economy bridges, team-escrow tools, missing disabled categories
- [x] 3.5 Extension module: use `appinit.ProvidesMCP`/`appinit.ProvidesObservability`, add MCP mgmt to catalog, add lifecycle ComponentEntry for mcp-manager and observability-token-cleanup
- [x] 3.6 Verify build + tests: `go build ./internal/app/...` && `go test ./internal/app/...`

## 4. app.New() Rewrite (Fix 2 — Steps 2.4-2.5)

- [x] 4.1 Rewrite `app.New()` — Phase A (module build) + Phase B (post-build wiring)
- [x] 4.2 Extract helper: `populateAppFields(app, resolver)` — resolver → app field mapping
- [x] 4.3 Extract helper: `buildCatalogFromEntries(entries)` — CatalogEntry → Catalog
- [x] 4.4 Extract helper: `buildHookRegistry(cfg, bus)` — hook registry construction
- [x] 4.5 Extract helper: `buildApprovalProvider(cfg, gw)` — approval provider + grant store
- [x] 4.6 Extract helper: `wirePostAgent(app, resolver, tools, bus, ...)` — A2A, P2P executor, routes, audit
- [x] 4.7 Extract helper: `registerPostBuildLifecycle(app)` — gateway + channels only
- [x] 4.8 Extract helper: `wireMemoryAndTurnCallbacks(app, iv, fv)` — compaction + turn triggers
- [x] 4.9 Verify build + tests: `go build ./internal/app/...` && `go test ./internal/app/...`

## 5. Dead Code Removal (Fix 2 — Step 2.6)

- [x] 5.1 Delete `registerLifecycleComponents()` from `app.go`
- [x] 5.2 Delete `registerEconomyLifecycle()` from `wiring_economy.go`
- [x] 5.3 Delete `registerObservabilityLifecycle()` from `wiring_observability.go`
- [x] 5.4 Verify build + tests: `go build ./...` && `go test ./internal/app/...`

## 6. Downstream Sync (Fix 3)

- [x] 6.1 Add OutputManager section to `docs/configuration.md` (tokenBudget, headRatio, tailRatio)
- [x] 6.2 Fix Presidio URL default (`http://localhost:5002`)
- [x] 6.3 Fix Session maxHistoryTurns default (`50`), Exec defaultTimeout (`30s`), Filesystem maxReadSize (`10MB`)
- [x] 6.4 Fix P2P enableRelay default (`true`), toolIsolation.maxMemoryMB (`256`)

## 7. Final Verification

- [x] 7.1 `go build ./...` — full project build passes
- [x] 7.2 `go test ./...` — all tests pass (except pre-existing `deadline` timing flake)
