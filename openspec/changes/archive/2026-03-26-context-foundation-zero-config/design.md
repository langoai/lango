## Context

Lango's context engineering stack (Knowledge, Observational Memory, Embedding/RAG, Graph Store, Librarian) exposes 26+ individual config fields across `KnowledgeConfig`, `ObservationalMemoryConfig`, `EmbeddingConfig`, `GraphConfig`, `LibrarianConfig`, and `SkillConfig`. Each subsystem defaults to `Enabled: false` (plain `bool`), so new users must manually enable and tune every subsystem. When a dependency is missing (e.g., no embedding provider), the feature is silently skipped with only an Info log — invisible in TUI or CLI output.

Current diagnostic tools (doctor, status) check subsystems independently but have no unified "context health" view. The status command hardcodes 16 features in `collectFeatures()` with no reason/suggestion metadata. The doctor system has 22 checks but no aggregated context-level diagnostic.

The config loader (`config.Load()`) returns `*Config` and uses `viper.SetDefault()` before `ReadInConfig()`, making `viper.IsSet()` unreliable for distinguishing user-set values from defaults — a `bool` field set to `false` by default looks identical to one explicitly set to `false` by the user.

## Goals / Non-Goals

**Goals:**
- Single-line context configuration via `contextProfile: balanced`
- User explicit overrides are never silently replaced by profile defaults
- Structured diagnostics: every context subsystem reports why it's off and what to do
- Doctor check aggregates context subsystem health into one actionable view
- Status command shows profile name and per-feature reason

**Non-Goals:**
- Auto-detection of embedding providers (deferred to separate change)
- Auto-enable features when dependencies become available (deferred)
- TUI settings form for context profile (can follow separately)
- Changes to how subsystems actually work (this is config + diagnostics only)

## Decisions

### D1: Raw viper instance for explicit key detection

**Decision:** Use a second, defaults-free `viper.Viper` instance to read the same config file and detect which keys the user explicitly set. Return `*LoadResult{Config, ExplicitKeys}` from `Load()`.

**Alternatives considered:**
- `*bool` tri-state fields: Would require migrating 6+ config fields from `bool` to `*bool`, updating all consumers, and changing serialization. High churn for a narrow benefit.
- Viper `IsSet()` on the main instance: Fails because `SetDefault()` marks all defaulted keys as "set." Verified locally.
- Config overlay (file-first, then defaults): Would require rewriting the loader's merge strategy. Viper doesn't support this natively without two-pass reading.

**Rationale:** The raw viper approach is additive (no field type changes), correct (only file-present keys return true), and contained (only `Load()` changes internally). The `LoadResult` return type is a compile-safe breaking change — all callers fail at build time, not runtime.

### D2: FeatureStatus in `internal/types/`

**Decision:** Place `FeatureStatus` in `internal/types/feature_status.go`. Adapters live in their CLI packages (`cli/status/adapter.go`, `cli/doctor/checks/adapter.go`).

**Alternatives considered:**
- `internal/app/wiring_status.go`: Would require app → CLI imports for adapters (layer violation).
- New `internal/featurestatus/` package: Overengineered for a single struct.

**Rationale:** `internal/types/` is the existing shared-types package. Adapters in CLI layer respect the dependency direction (CLI → types, not app → CLI).

### D3: Profile applies before PostLoad, not inside it

**Decision:** Call `ApplyContextProfile(cfg, explicitKeys)` inside `Load()` after `Unmarshal` but before `PostLoad(cfg)`. `PostLoad`'s signature stays unchanged.

**Rationale:** Profile application must happen before `Validate()` (which is called by `PostLoad`) so that profile-set values are validated. Keeping `PostLoad` unchanged minimizes the breaking surface — only `Load()` return type changes.

### D4: Doctor check, not standalone tool

**Decision:** Implement context diagnostics as a `checks.Check` registered in `AllChecks()`, not as an agent tool.

**Rationale:** Doctor checks are the established pattern for user-facing diagnostics (22 existing checks). Agent tools are for runtime operations. Context health is a config-time diagnostic — doctor is the right home.

### D5: StatusCollector in app layer

**Decision:** `StatusCollector` lives in `internal/app/wiring_status.go`, collects `types.FeatureStatus` from wiring functions, and is passed to the status command and doctor check via bootstrap.

**Rationale:** Wiring functions are the natural point where init success/failure is determined. The collector aggregates results without adding new interfaces to the wiring functions — they just return an extra `*types.FeatureStatus` value.

## Risks / Trade-offs

- **[Load() signature break]** → All callers must adapt. Mitigated by compile-time safety — `*LoadResult` is not assignable to `*Config`. Grep for `config.Load(` to find all callsites.
- **[Profile doesn't cover all 26 knobs]** → Profile sets only the 6 context-related `Enabled` flags. Numeric thresholds (e.g., `messageTokenThreshold`) keep their defaults. This is intentional — profiles control "what's on," not "how it's tuned."
- **[ExplicitKeys is nil when no config file exists]** → Handled: `ApplyContextProfile` treats nil as "no explicit overrides" — profile applies fully, which is the desired behavior for new users.
- **[StatusCollector adds a new return value to wiring functions]** → 5 functions change signature. Low risk since these are internal and called only in `app.New()`.
