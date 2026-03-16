## Context

The config/bootstrap path currently scatters normalization and validation across multiple call sites: `Load()` runs all 5 steps, but `configstore.Store.Load()` (via bootstrap) returns raw deserialized config without normalization, `Store.Save()` only calls `Validate()` (no path normalization), and the CLI `config set` command bootstraps twice — once for the loader, once for the saver — causing keyfile shred on the first bootstrap to fail the second. The collaborator preset also enables `payment.enabled=true` without setting `rpcUrl`, causing `Validate()` to fail.

## Goals / Non-Goals

**Goals:**
- Single `PostLoad()` function for all normalization+validation, reusable across Load, bootstrap, and Save paths
- Eliminate double-bootstrap in `config set` via cleanup-function pattern
- Ensure `Store.Save()` persists only canonical (normalized+validated) configs
- Fix collaborator preset to pass validation with correct RPC URL for Base Sepolia

**Non-Goals:**
- Refactoring the entire bootstrap pipeline
- Changing the encrypted config storage format
- Modifying other CLI commands beyond `config set`

## Decisions

### Decision 1: `PostLoad()` as single entry point
**Choice**: Export a `PostLoad(*Config) error` function that chains `MigrateEmbeddingProvider` → `substituteEnvVars` → `NormalizePaths` → `ValidateDataPaths` → `Validate`.

**Rationale**: All 5 operations are idempotent. Grouping them ensures no call site can forget a step. `Load()` delegates to it; `phaseLoadProfile()` calls it once at the end; `Store.Save()` calls it before marshal.

**Alternative**: Keep separate functions and document the call order — rejected because the current state already shows that distributed responsibility leads to regressions.

### Decision 2: Cleanup-function pattern for config set
**Choice**: Change `cfgLoader` signature from `func() (*Config, error)` to `func() (*Config, func(), error)`. The cleanup function closes DBClient via `defer` in `RunE`.

**Rationale**: Cobra's `PostRunE` does not execute when `RunE` returns an error, making it unreliable for resource cleanup. The closure pattern ensures cleanup runs on all code paths (success, error, panic).

**Alternative**: Use a shared bootstrap result variable with manual Close() — rejected because it requires careful ordering and is error-prone.

### Decision 3: PostLoad in Store.Save()
**Choice**: Call `PostLoad()` at the start of `Save()`, mutating the config before marshaling.

**Rationale**: Guarantees the persisted form is always canonical. Double-calling PostLoad is safe because all operations are idempotent (absolute paths stay absolute, expanded env vars have no patterns, already-migrated fields are no-ops, Validate is a pure check).

## Risks / Trade-offs

- **[Double PostLoad calls]** → Some paths (e.g., `MigrateFromJSON`) will call PostLoad twice (once in Load, once in Save). Mitigated by ensuring all operations are idempotent. The cost is negligible CPU.
- **[Breaking cfgLoader signature]** → `NewSetCmd` callers must update. Mitigated by this being internal API with a single call site in `main.go`.
- **[Save mutates config]** → `PostLoad()` in `Save()` modifies the config in-place before persisting. This is intentional — callers should expect the saved config to be in canonical form. Documented in the function comment.
