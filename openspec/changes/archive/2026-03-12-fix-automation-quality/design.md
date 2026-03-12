## Context

The automation subsystems (cron scheduler, background manager, workflow engine) were recently added with core features: per-job timeout, idempotent upsert, one-time job support, and shutdown-aware semaphores. Code review revealed five quality issues ranging from potential panics to unnecessary DB calls.

## Goals / Non-Goals

**Goals:**
- Eliminate double-close panic in `Scheduler.Stop()`
- Reduce DB round-trips in `AddJob` by returning the persisted job from `Upsert`
- Remove dead/redundant code to improve maintainability
- Fix import ordering to match Go style guidelines

**Non-Goals:**
- Adding application-level locking for Upsert (DB UNIQUE constraint is sufficient)
- Adding shutdown timeout to workflow engine (context cancellation handles this)
- Refactoring background manager's semaphore timing (existing behavior is acceptable)

## Decisions

### Decision 1: sync.Once for Stop() instead of boolean flag

Use `sync.Once` to protect `Scheduler.Stop()` from double-close panics on `shutdownCh`.

**Rationale**: `sync.Once` is the idiomatic Go pattern for exactly-once execution. A boolean flag + mutex would work but adds more fields and is error-prone. `sync.Once` is goroutine-safe and self-documenting.

### Decision 2: Return `*Job` from Upsert instead of adding a cache

Change `Store.Upsert` from `(bool, error)` to `(*Job, bool, error)` so `AddJob` can use the returned job directly without a follow-up `GetByName` call.

**Rationale**: On the update path, the job is already in memory after the `GetByName` + `Update` calls within Upsert. On the create path, a read-back is needed to get the generated UUID, but this replaces the external `GetByName` in `AddJob` — same total DB calls for create, one fewer for update.

**Alternatives considered**: Caching the last-upserted job was rejected as overengineering for this low-frequency operation.

### Decision 3: Remove redundant unregisterJob rather than document it

The `disableOneTimeJob` method called `unregisterJob`, but this was already done by the `sync.Once` wrapper in `registerJob`. Rather than adding a comment explaining why it's a no-op, remove the redundant call entirely.

**Rationale**: A no-op call that looks intentional is confusing for future maintainers. The `sync.Once` wrapper in `registerJob` is the canonical location for unregistration of one-time jobs.

## Risks / Trade-offs

- **[Store interface breaking change]** → Only internal consumers exist (`EntStore`, two test mocks). All updated in this change. No external implementations known.
- **[sync.Once prevents re-start]** → After `Stop()`, the scheduler cannot be restarted because `shutdownCh` remains closed and `stopOnce` won't fire again. This is acceptable because the scheduler lifecycle is tied to the application lifecycle — restart requires a new `Scheduler` instance.
