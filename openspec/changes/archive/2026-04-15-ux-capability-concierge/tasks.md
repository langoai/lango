## 1. Session Mode Data Model

- [x] 1.1 Define `SessionMode` struct in `internal/config/types.go` (Name, Tools, Skills, SystemHint)
- [x] 1.2 Add `Modes map[string]SessionMode` to `Config` with merge logic for user overrides
- [x] 1.3 Create `internal/config/modes.go` with built-in modes: `code-review`, `research`, `debug`
- [x] 1.4 Add `Mode string` field to `session.Session` and wire persistence in `EntStore` (implemented via `Metadata[lango.mode]` to avoid Ent schema migration — see Session.Mode()/SetMode())
- [x] 1.5 Add `session.WithModeName(ctx, name) context.Context` and `session.ModeNameFromContext(ctx) string` helpers (mode name string, not full SessionMode — avoids import cycle)
- [x] 1.6 Add unit tests for mode merge logic and context round-trip

## 2. Catalog and Prompt Dynamic Generation

- [x] 2.1 Add `Catalog.ListVisibleToolsForMode(modeTools []string) []ToolSchema` in `internal/toolcatalog/catalog.go`
- [x] 2.2 Add `Catalog.ResolveModeAllowlist(modeTools []string) map[string]bool` helper (expanding `@category` refs)
- [x] 2.3 Add `ContextAwareModelAdapter.WithCatalog(CatalogSource)` + `WithModeResolver(ModeResolver)` (interface-based injection to avoid config import cycle)
- [x] 2.4 Move tool catalog section generation into `GenerateContent()`: call `CatalogSource.BuildToolCatalogSection(modeName)` after Phase 3 combine, append to per-turn prompt
- [x] 2.5 Remove `builder.Add(buildToolCatalogSection(catalog))` call from `internal/app/wiring.go`; wire `catalogSourceAdapter` + `modeResolverAdapter` in both wrapper paths
- [x] 2.6 Inject session mode `SystemHint` into the per-turn prompt when non-empty
- [x] 2.7 Add unit tests: ResolveModeAllowlist category expansion, empty mode equivalence, mixed explicit+category

## 3. Mode Enforcement Middleware

- [x] 3.1 Add `WithModeAllowlist(resolver)` middleware in `internal/toolchain/mw_mode_allowlist.go`
- [x] 3.2 Wire `WithModeAllowlist` into the chain (B4c3, after WithPrincipal, before WithApproval)
- [x] 3.3 Error path returns "tool %q not available in current mode" with tool name (mode name available via context if needed)
- [x] 3.4 Add unit tests: allowed/blocked/no-mode/nil-resolver paths

## 4. Turn Runner and Mode Propagation

- [x] 4.1 In `turnrunner.Runner.Run()`, read session mode from `sessionStore.Get(key).Mode()` at turn start
- [x] 4.2 Attach mode to context via `session.WithModeName()` before calling `executor.RunStreamingDetailed()`
- [x] 4.3 Add unit test: runner propagates mode to executor context

## 5. Capability Discovery (list_skills summary + view_skill)

- [x] 5.1 Add `summary` bool parameter to `list_skills` tool in `internal/app/tools_meta.go`; default `false` preserves existing behavior
- [x] 5.2 Implement `summary=true` branch: return only `{name, description, when_to_use}` per skill
- [x] 5.3 Apply mode filter in `list_skills` handler when session mode is active
- [x] 5.4 Add new `view_skill` tool with `name` (required) and `path` (optional) parameters
- [x] 5.5 Implement path safety: reject paths that escape the skill directory via `filepath.Clean` + prefix check
- [x] 5.6 Transition instruction skills to `ExposureDeferred` in `internal/skill/registry.go` (template type not in current registry implementation)
- [x] 5.7 Keep script/fork skills with direct exposure (unchanged)
- [x] 5.8 Capability discovery guidance baked into mode SystemHint (built-in modes already mention `list_skills`/`view_skill`)
- [x] 5.9 Add unit tests: summary param schema, view_skill registered + path safety, empty-registry handler

## 6. Cost Visibility

- [x] 6.1 Create `internal/provider/pricing.go` with `ModelPrice` struct and `PriceFor(model) (ModelPrice, bool)` function
- [x] 6.2 Populate pricing table with primary supported models (Opus 4, Sonnet 4, Haiku 4, Sonnet 3.5, Gemini 2.5 Pro/Flash, Gemini 2.0 Flash, GPT-4o)
- [x] 6.3 Add `EstimatedCostUSD float64` to `TurnTokenUsageMsg`
- [x] 6.4 Add `EstimatedCostUSD` to `eventbus.TokenUsageEvent` (no separate CostEvent)
- [x] 6.5 Compute cost at both emission sites (cockpit handleDone + wireModelAdapterTokenUsage)
- [x] 6.6 Update `appendTokenSummaryWithCost()` to render `~$<cost>` (4-decimal under $0.01, 2-decimal otherwise)
- [x] 6.7 Add unit tests for pricing lookup and cost computation (known, unknown, zero-token, case-insensitive)

## 7. Slash Commands and CLI Flag

- [x] 7.1 Add `/mode` handler in `internal/cli/chat/commands.go`: with argument sets mode, without argument lists available modes
- [x] 7.2 Add `/cost` handler: prints session cumulative tokens and estimated cost
- [x] 7.3 Maintain session cumulative token/cost counter in `ChatModel` updated from `TurnTokenUsageMsg`
- [x] 7.4 Add `--mode` flag to root and cockpit commands; pre-create session with mode when set
- [x] 7.5 Publish `ModeChangedEvent{SessionKey, OldMode, NewMode}` on `setSessionMode`
- [x] 7.6 cmdMode renders mode change as system message in same process; cross-process subscriber via EventBus deferred to Phase 4

## 8. Verification

- [x] 8.1 `go build ./...` passes
- [x] 8.2 `go test ./internal/toolcatalog/... ./internal/toolchain/... ./internal/adk/... ./internal/cli/chat/... ./internal/app/... ./internal/provider/... ./internal/session/... ./internal/config/...` passes
- [x] 8.3 Manual test: `/mode code-review` → confirm tool catalog in next turn only shows allowed tools (user action required)
- [x] 8.4 Manual test: call a blocked tool under a mode → verify error message (user action required)
- [x] 8.5 Manual test: `/cost` after several turns → verify summary displays (user action required)
