## Context

The tool subsystem initialization in `internal/app/app.go` and associated wiring files follows a pattern: each subsystem checks if it's enabled, creates components, registers tools with the catalog, and optionally registers lifecycle components. However, audit reveals that only 4 of 17+ subsystems register disabled categories, only SmartAccount has a disabled-state message (and it's incomplete), several subsystems silently skip initialization without warning logs, and SessionGuard has no Stop() method causing goroutine leaks.

## Goals / Non-Goals

**Goals:**
- Every subsystem registers a disabled category when off, so `builtin_health` can report full system state
- Config validation catches missing required fields before runtime failures
- All lifecycle-aware components are registered for graceful shutdown
- Warning logs emitted when subsystems degrade or skip initialization

**Non-Goals:**
- Refactoring the overall wiring architecture (e.g., replacing with appinit modules)
- Adding new tool capabilities or changing tool behavior
- Modifying the tool catalog API itself

## Decisions

1. **Config validation via Validate() method on config types** — Keeps validation co-located with the type definition. The alternative (validating in wiring functions) scatters validation logic and leads to app/CLI divergence (bug A6).

2. **Disabled category registration inline with if/else** — Each subsystem's enabled block gets a corresponding else block registering the disabled category. The alternative (a centralized registration block at the end) would duplicate config-key knowledge and be harder to maintain as subsystems are added.

3. **SessionGuard.Stop() sets active=false** — The guard already has a mutex-protected `active` field and `Start()` sets it to true. Adding `Stop()` that sets it to false is the minimal change. The event bus subscription remains (no Unsubscribe API), but handleAlert returns early when inactive.

4. **Warning logs use logger().Warn with structured fields** — Consistent with existing codebase patterns. Includes `"fix"` field where actionable to guide operators.

## Risks / Trade-offs

- [Risk] Disabled category registration increases catalog size → Minimal memory impact, and the diagnostic value far outweighs it
- [Risk] SessionGuard subscription remains after Stop() → Bus callback is a no-op (early return), negligible overhead. Full unsubscribe would require eventbus API changes out of scope
- [Risk] Validate() only checks 3 fields → Sufficient for the critical runtime failures identified. Can be extended later without breaking changes
