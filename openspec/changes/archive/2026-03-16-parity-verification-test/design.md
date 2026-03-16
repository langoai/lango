## Context

`app.New()` was refactored from a monolithic 900-line function into a 5-module system using `appinit.Builder.Build()`. The original migration plan included a parity gate before dead code removal, but this step was skipped. The current codebase lacks tests that exercise the full `app.New()` path and verify the resulting application state matches expectations.

The test infrastructure (`testutil.TestEntClient`) provides in-memory SQLite + ent clients suitable for integration tests without external dependencies.

## Goals / Non-Goals

**Goals:**
- Verify `buildCatalogFromEntries()` correctly registers categories, tools, and enabled/disabled states
- Verify `registerPostBuildLifecycle()` registers gateway and channel components with correct names
- Verify `app.New()` with default config produces expected catalog categories, lifecycle components, and field population
- Verify `app.New()` with feature flags (knowledge, graph, memory, cron) enables the correct additional categories and components
- Provide a `Names()` accessor on `lifecycle.Registry` for test introspection

**Non-Goals:**
- Testing individual tool handler behavior (covered in tool package tests)
- Testing middleware chain ordering (covered in toolchain package tests)
- Testing P2P/Payment/MCP enabled scenarios (require external dependencies)
- Performance benchmarking of `app.New()`

## Decisions

### Two-layer test structure
Layer 1 tests pure helper functions without any infra (no DB, no bootstrap). Layer 2 tests call real `app.New()` with `testutil.TestEntClient` fixtures. This separation ensures fast CI feedback from Layer 1 while Layer 2 catches integration regressions.

**Alternative considered**: Single integration test only. Rejected because helper function bugs would be harder to isolate.

### Name-level verification (not deep equality)
Tests verify category names, enabled flags, lifecycle component names, and field nil/non-nil status — not deep structural equality. This avoids brittle tests that break on unrelated changes (e.g., adding a new tool to an existing category).

**Alternative considered**: Snapshot-based golden file tests. Rejected because they would need frequent updates and obscure the intent.

### `RawDB: nil` in fixtures
`bootstrap.Result.RawDB` is only used by `initEmbedding()` for sqlite-vec. Since test configs leave `Embedding.Provider` empty, `initEmbedding()` returns nil immediately, making `RawDB: nil` safe.

## Risks / Trade-offs

- [Risk] Tests depend on `config.DefaultConfig()` defaults → If defaults change, tests may need updates → **Mitigation**: Tests assert minimum expectations (e.g., `ToolCount >= 11`) rather than exact values
- [Risk] Agent creation requires a supervisor which requires provider config → **Mitigation**: Empty provider still creates a valid supervisor; `initAgent` handles nil model gracefully
- [Trade-off] Layer 2 tests are slower (~3s) due to ent client setup → Acceptable for CI; skippable with `-short` flag
