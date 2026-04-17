## Context

Three P1 features need data before design decisions:
1. **Shared task coordination** — is broadcast delegation wasteful?
2. **Child session reset** — do sessions accumulate context bloat?
3. **Workspace cleanup** — do cleanups silently fail?

All three already have the right hooks/events in place. The gap is observability.

## Goals / Non-Goals

**Goals:**
- Make team task delegation metrics visible in logs
- Make child session lifecycle transitions visible in logs with timestamps
- Make workspace cleanup failures visible in logs

**Non-Goals:**
- Persistent metrics storage (counters are in-memory, lost on restart — sufficient for initial observation)
- New CLI commands or doctor checks for these metrics
- Any behavioral changes

## Decisions

### D1: Team task metrics via EventBus subscriber

**Choice**: New `bridge_team_metrics.go` in `internal/app/` that subscribes to `TeamTaskDelegatedEvent` and `TeamTaskCompletedEvent`. Logs at Info level with structured fields. Maintains in-memory counters (total delegations, total workers dispatched, duplicate work ratio proxy).

**Why not persist?** The goal is to observe over a few operational sessions. Structured logs are greppable. If the signal is strong enough for P1 implementation, we'll add persistence then.

### D2: Child lifecycle logging at Info level

**Choice**: In `wiring.go:703` childHook, add `logger().Infow(...)` for each event type alongside the existing provenance calls. Include `childKey`, `parentKey`, `agentName`, and `time.Now()`.

**Why Info not Debug?** Current Debug-level error logging is only on failures. Normal lifecycle transitions (fork/merge/discard) are invisible. Info makes them greppable without enabling Debug globally.

### D3: Workspace cleanup error logging

**Choice**: Replace `_ = m.RemoveWorktree(path)` and `_ = m.DeleteBranch(branch)` with `if err := ...; err != nil { log.Warnw(...) }`. Existing `logging` package used.

**Why Warn?** Cleanup failure doesn't block the current operation but indicates resource leak. Warn is the correct severity — visible in default log level, not noisy.
