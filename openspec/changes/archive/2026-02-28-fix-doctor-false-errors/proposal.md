## Why

`lango doctor` reports false errors for configurations that have working runtime defaults. Graph Store reports error when `graph.databasePath` is empty (but wiring.go creates a fallback path automatically). Observational Memory reports error when `maxMessageTokenBudget` is 0 (but `DefaultConfig()` never initialized it). The database check also uses a stale `sessions.db` fallback instead of the current `lango.db` convention. These false errors confuse users who have a working setup.

## What Changes

- Add `ObservationalMemory` defaults to `DefaultConfig()` and viper defaults (messageTokenThreshold=1000, observationTokenThreshold=2000, maxMessageTokenBudget=8000, etc.)
- Downgrade Graph Store `databasePath` empty check from error to warning (runtime fallback exists in wiring.go)
- Fix database doctor check fallback path from `sessions.db` to `lango.db`
- Update test and spec references from `sessions.db` to `lango.db`

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `cli-doctor`: Downgrade graph store databasePath check from error to warning; fix database check fallback path from sessions.db to lango.db
- `config-system`: Add ObservationalMemory defaults to DefaultConfig() and viper SetDefault calls

## Impact

- `internal/config/loader.go` — DefaultConfig() and viper defaults
- `internal/cli/doctor/checks/graph_store.go` — severity change
- `internal/cli/doctor/checks/database.go` — fallback path fix
- `internal/cli/doctor/checks/checks_test.go` — test update
- `openspec/specs/server/spec.md` — spec reference update
