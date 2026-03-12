## Why

The automation subsystems (cron, background, workflow) have several quality and correctness issues discovered during code review: double-close panic in Scheduler.Stop(), redundant DB queries in AddJob, dead code in test mocks, import ordering violations, and a confusing redundant unregisterJob call. These issues risk production panics and unnecessary database load.

## What Changes

- **Scheduler.Stop() double-close protection**: Wrap Stop() with `sync.Once` to prevent panic when called multiple times (e.g., during graceful shutdown sequences)
- **Upsert returns persisted Job**: Change `Store.Upsert` signature from `(bool, error)` to `(*Job, bool, error)` to eliminate the redundant `GetByName` call in `AddJob` — **BREAKING** (Store interface change)
- **Remove redundant unregisterJob**: The `disableOneTimeJob` method called `unregisterJob` which was already called by the `sync.Once` wrapper in `registerJob`
- **Remove unused test mock field**: `mockAgentRunner.delay` field and its associated sleep logic were dead code
- **Fix import ordering**: Move `"time"` import into the stdlib group in `tools_automation.go`

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `cron-scheduling`: Upsert now returns the persisted `*Job`, Stop() is idempotent via sync.Once, disableOneTimeJob no longer redundantly calls unregisterJob

## Impact

- **internal/cron/store.go**: `Store` interface change — `Upsert` signature updated (breaking for any external implementations)
- **internal/cron/scheduler.go**: `AddJob` simplified (no more `GetByName` after Upsert), `Stop()` wrapped in `sync.Once`, `disableOneTimeJob` simplified
- **internal/cron/scheduler_test.go**: `mockStore.Upsert` updated, `mockAgentRunner.delay` removed
- **internal/testutil/mock_cron.go**: `MockCronStore.Upsert` updated
- **internal/app/tools_automation.go**: Import ordering fix
