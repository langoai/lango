## Context

`lango doctor` reports false errors for two checks that have working runtime defaults:
1. **Graph Store**: `graph.databasePath` empty → StatusFail, but `wiring.go` creates a fallback path (`graph.db` next to session DB)
2. **Observational Memory**: `maxMessageTokenBudget` is 0 → StatusFail, but `DefaultConfig()` never initialized `ObservationalMemory` at all (despite struct comments claiming defaults of 1000/2000/8000)

Additionally, `database.go` uses a stale `sessions.db` fallback path instead of the current `lango.db` convention established in the onboard wizard UX fix.

## Goals / Non-Goals

**Goals:**
- Eliminate false error reports from `lango doctor` for working configurations
- Add missing `ObservationalMemory` defaults to `DefaultConfig()` and viper
- Align database check fallback path with current `lango.db` convention

**Non-Goals:**
- Changing runtime wiring behavior (it already works correctly)
- Adding new doctor checks
- Changing the actual graph store fallback logic in wiring.go

## Decisions

1. **Add ObservationalMemory to DefaultConfig()** — The struct comments document defaults (1000/2000/8000/5/20/4000/5) but no code enforces them. Fix at the source by adding defaults to `DefaultConfig()` and registering viper defaults. This ensures the values are always populated regardless of config file contents.

2. **Downgrade graph.databasePath check to warning** — Since `wiring.go:initGraphStore()` already handles empty `DatabasePath` by deriving a path from session DB location, this is not an error. Change to StatusWarn with an informational message explaining the fallback.

3. **Fix database.go fallback path** — Simple find-and-replace: `sessions.db` → `lango.db` in the `resolveDatabasePath()` fallback. The `DefaultConfig()` already uses `lango.db`.

## Risks / Trade-offs

- [Reduced strictness] Users with genuinely misconfigured graph paths will see a warning instead of an error → Acceptable because the runtime handles it gracefully.
- [New defaults may override user intent] If a user explicitly sets `maxMessageTokenBudget: 0` intending to disable budget limits → Low risk since 0 was never a valid value (check always failed on it).
