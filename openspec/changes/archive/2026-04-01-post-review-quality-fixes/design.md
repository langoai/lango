## Context

The exec-safety-followup batch (6 changes) passed all tests but code review found 3 correctness/security issues and 1 concurrency gap. All fixes are already implemented; this design documents the decisions made.

## Goals / Non-Goals

**Goals:**
- Budget persistence reflects actual per-run state, not stale shared state
- env wrapper unwrap handles POSIX + GNU coreutils flag syntax
- Policy observability works in default configuration without hook system
- Concurrent sessions don't corrupt each other's budget stats

**Non-Goals:**
- Shared backoff package consolidation (deferred)
- CommandGuard API restructuring for stringly-typed kill detection (deferred)
- Recovery API redesign for Decide() mutation side-effect (deferred)

## Decisions

**D1: Session-local budget via `sessionBudgetState` + `LastRunStatsForSession`**
Per-session cumulative counters tracked in `budgetRestoringExecutor.sessionState` (sync.Map). `CoordinatingExecutor` stores run stats keyed by sessionID in a `sync.Map` (not a single slot), retrieved via `LastRunStatsForSession` with consume-once (`LoadAndDelete`) semantics. This avoids process-shared budget contamination and concurrent session stats overwrite.

**D2: `skipEnvArgs` with strict assignment detection**
Env arguments parsed per POSIX/GNU syntax: standalone flags (`-i`, `-0`), flag-with-argument (`-u NAME`, `-C DIR`, `-S STRING`), terminator (`--`), and variable assignments validated via `looksLikeEnvAssignment` (shell variable name pattern before `=`). Rejects paths (`./foo=bar`) and flags (`--flag=val`).

**D3: Policy bus decoupled from hook gate**
`policyBus` initialized when `bus != nil`, regardless of `cfg.Hooks.EventPublishing`. Policy events are a distinct concern from hook tool-execution events.

**D4: `initAgentRuntime` return reverted**
Process-shared budget no longer used for serialization, so exposing it via return value is unnecessary. Reverted to single `turnrunner.Executor` return.

## Risks / Trade-offs

- [Risk] `sync.Map` in `runStatsMap` grows unbounded → Mitigated by `LoadAndDelete` (entries consumed after each run)
- [Risk] `skipEnvArgs` may not cover all GNU env extensions → Covered common flags; unknown flags treated as standalone (safe fallback)
- [Risk] Unconditional policy bus increases event volume → Negligible; only observe/block verdicts are published (allow is skipped)
