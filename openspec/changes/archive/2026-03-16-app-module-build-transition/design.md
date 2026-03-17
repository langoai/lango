## Context

`app.New()` was a 900+ line monolithic initializer that performed all component initialization sequentially. The `appinit` module system (Phase 1-3: Module interface, Builder, TopoSort) was implemented previously but `modules.go` was dead code — `app.New()` never called `builder.Build()`. Additionally, OpenSpec specs had drifted from code, and docs had undocumented config defaults.

## Goals / Non-Goals

**Goals:**
- Transition `app.New()` from monolithic sequential init to `appinit.Builder.Build()` with 5 modules
- Achieve 100% behavioral parity with the old `app.New()` (same tools, catalog entries, lifecycle components)
- Fix spec/code alignment for `config-default-walker` and `cli-bootstrap-factory`
- Sync downstream docs with actual config defaults

**Non-Goals:**
- No new features or config keys
- No changes to CLI commands or external API
- No changes to module dependency graph (already correct from Phase 1-3)

## Decisions

### D1: Module lifecycle components via ComponentEntry return (not direct registry calls)
Each module returns `[]lifecycle.ComponentEntry` in `ModuleResult`. Post-build, `app.New()` registers all returned entries with the lifecycle registry. This replaces direct `app.registry.Register()` calls scattered across modules.

**Rationale**: Modules become self-contained units — they declare what they need started/stopped. The app layer just collects and registers. This eliminates the need for modules to hold a `*lifecycle.Registry` reference.

**Alternative considered**: Keep direct registry calls in modules → rejected because it couples modules to app internals and makes testing harder.

### D2: Post-build wiring phase for cross-cutting concerns
After `builder.Build()`, `app.New()` applies cross-cutting middleware (WithLearning, WithOutputManager, WithHooks, WithApproval) and creates the agent. This cannot be done inside modules because middleware must wrap ALL tools from ALL modules.

**Rationale**: Middleware ordering is critical (learning → output → hooks → approval). A centralized post-build phase ensures correct ordering regardless of module init order.

### D3: CatalogEntries in BuildResult
`BuildResult` collects `CatalogEntries` from all modules. Post-build, `buildCatalogFromEntries()` converts them into a `toolcatalog.Catalog`. Dispatcher tools are added after catalog construction.

**Rationale**: Modules declare their catalog categories; the app layer builds the unified catalog. This avoids modules needing direct catalog access.

### D4: Spec follows code for config-default-walker
The spec was updated to match the existing unexported `setDefaultsFromStruct` implementation rather than forcing code to match the spec's exported `WalkDefaults` signature.

**Rationale**: The current code design is superior — viper defaults are set directly without an intermediate map, and maps are correctly skipped (they contain dynamic user content).

### D5: Code follows spec for cli-bootstrap-factory
`serveCmd()` was changed to use `cliboot.BootResult()` instead of direct `bootstrap.Run()`.

**Rationale**: The spec requirement ("no direct bootstrap calls in cmd/") is correct — centralizing bootstrap through `cliboot` prevents divergence.

## Risks / Trade-offs

- **[Risk] Behavioral divergence during transition** → Mitigated by 1:1 comparison of each module's Init() against the corresponding section of old `app.New()`. All existing tests pass.
- **[Risk] Lifecycle ordering change** → Module components are registered in module order (foundation → intelligence → automation → network → extension), then gateway + channels. This matches the old ordering. Priority-based sorting in the registry ensures correct start/stop order.
- **[Trade-off] populateAppFields() uses type assertions** → The `Resolver` returns `interface{}`, requiring type assertions in `populateAppFields()`. This is inherent to the module system's decoupled design. The alternative (typed resolver) would create import cycles.
