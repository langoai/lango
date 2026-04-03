## Context

Phases 1-3 hardened security defaults, added integration tests, and wired Prometheus/OTel/alerting. The public surface (README, installation docs, TUI settings) still reflects pre-Phase-1 positioning. New users see "sovereign economic stack" before "trustworthy runtime", cannot tell which TUI settings are experimental, and lack platform-specific CGO guidance.

## Goals / Non-Goals

**Goals:**
- Shift README messaging from economy-first to trust-first positioning
- Make feature maturity visible at the TUI settings level via `[EXP]` badges
- Provide platform-specific C compiler installation guidance for CGO
- Reflect Phase 1-3 completion in the roadmap document

**Non-Goals:**
- Structural module consolidation (deferred — `llm`, `ctxkeys`, `toolparam` all have spec or cycle constraints)
- Homebrew tap distribution (gated on external `langoai/homebrew-tap` repo creation)
- Individual form title modifications (menu badge is the single source of truth for maturity display)
- New spec-level behavioral requirements (this is presentation/docs only)

## Decisions

### 1. Menu badge over form title churn

**Decision:** Add `[EXP]` badge via `renderItem()` in menu.go using an `ExperimentalCategories` map, rather than modifying 20+ individual form title strings.

**Rationale:** The existing badge system (`ADV`, dependency warnings) provides a proven pattern. A centralized map is a single point of update when features graduate from experimental. Individual form title modifications would require touching 13+ form files and their associated tests.

**Alternative considered:** Per-form `[Experimental]` title prefix — rejected due to maintenance cost and test churn.

### 2. Exported ExperimentalCategories map with drift test

**Decision:** Export the map as `ExperimentalCategories` and add a sorted-slice equality test in `menu_test.go`.

**Rationale:** If a new settings category is added but not classified, the test will catch the drift. The map is exported so tests and potentially future code (e.g., settings export) can reference it.

### 3. README CLI condensation via docs link

**Decision:** Replace 180-line inline CLI reference with 8 key commands + link to `docs/cli/index.md`.

**Rationale:** `docs/cli/index.md` is the authoritative CLI reference (already maintained). Duplicating it in README creates sync drift. The 8 commands cover the most common first-run actions.

### 4. Early-stage note placement

**Decision:** Move the stability warning from line 32 (below "Why Lango?") to directly after the badge row (line 14), before the tagline.

**Rationale:** First-time readers see the warning before feature claims. The maturity source of truth remains `docs/features/index.md`, linked from the note.

### 5. Roadmap update style

**Decision:** Add an "Execution Progress" table and mark backlog items with status column, rather than rewriting the roadmap structure.

**Rationale:** The existing roadmap structure (workstreams, backlog, future tracks) is sound. Adding status tracking preserves the original planning intent while reflecting actual progress.

## Risks / Trade-offs

- **[ExperimentalCategories stale]** → Mitigated by drift-prevention test that fails when the set changes without updating both map and test
- **[README churn on merge]** → Low risk since changes are in distinct sections (tagline, warning, Why Lango order, CLI block). Each edit targets a unique string anchor.
- **[Installation doc platform coverage]** → Only covers macOS, Ubuntu/Debian, Fedora/RHEL, Alpine. Other distros must adapt. This covers the most common developer environments.
