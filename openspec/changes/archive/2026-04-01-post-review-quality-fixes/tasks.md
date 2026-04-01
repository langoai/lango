# Tasks

## Fix A: Session-local budget persistence
- [x] Add `RunStats` struct and `runStatsMap sync.Map` to `CoordinatingExecutor`
- [x] Replace `LastRunStats()` with `LastRunStatsForSession(sessionID)` using `LoadAndDelete`
- [x] Store run stats keyed by sessionID in `RunStreamingDetailed`
- [x] Add `sessionBudgetState` struct with per-session cumulative counters
- [x] Rework `budgetRestoringExecutor` to use session-local state
- [x] Update `wireSessionUsage` to serialize from session-local state
- [x] Revert `initAgentRuntime` return to single `Executor`
- [x] Update wiring in `app.go`
- [x] Add tests: baseline + accumulation, cross-session isolation, consume-once semantics

## Fix B: Env wrapper flag/variable skip
- [x] Add `skipEnvArgs` function handling -i, -0, -u NAME, -C DIR, -S STRING, --, NAME=value
- [x] Add `looksLikeEnvAssignment` with shell variable name validation
- [x] Add unwrap tests for env flags, assignments, -S, --, path rejection
- [x] Add policy evaluation tests for env-wrapped dangerous commands

## Fix C: Policy bus unconditional
- [x] Simplify policyBus gate to `bus != nil`
- [x] Verify policy integration tests pass

## Verification
- [x] `go build ./...` passes
- [x] `go test ./...` — 136 packages, 0 failures
