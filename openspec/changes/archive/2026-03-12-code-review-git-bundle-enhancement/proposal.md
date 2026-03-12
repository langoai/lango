## Why

Code review of the `feature/git-bundle-enhancement` branch identified 10 issues across code reuse, quality, and efficiency. Three critical bugs (context starvation, goroutine leak potential, subscription duplication), four major performance/maintainability issues, and three minor code quality issues need fixing to ensure production reliability.

## What Changes

- Fix health monitor context reuse: separate 10s timeout for ping vs git state collection
- Guard event subscriptions with `sync.Once` to prevent duplicate handlers on monitor restart
- Add `GetByToken()` to SessionStore for O(1) token lookup instead of O(N) linear scan
- Refactor cron `New()` from 6 positional params to `SchedulerConfig` struct
- Add `ListByStatusBefore()` to escrow Store interface for filtered DB queries
- Add 5s timeout to DID lookup during app initialization
- Extract shared `floatToMicroUSDC()` to eliminate duplicate conversion functions
- Add slice capacity hints in team tool builders
- Extract `runGit()` helper to deduplicate git command execution pattern

## Capabilities

### New Capabilities

(none — all fixes are implementation-level improvements to existing capabilities)

### Modified Capabilities

- `team-health-monitoring`: Health monitor ping/git-state context separation and subscription deduplication
- `cron-scheduling`: Scheduler constructor API changed to use config struct
- `economy-escrow`: Store interface extended with `ListByStatusBefore` method

## Impact

- **internal/p2p/team/health_monitor.go**: Context handling, sync.Once guard
- **internal/p2p/handshake/session.go**: New `GetByToken()` method
- **internal/app/app.go**: Session validator, DID timeout
- **internal/app/tools_team.go**: Slice capacity hints
- **internal/cron/scheduler.go**: `SchedulerConfig` struct, updated constructor
- **internal/cron/scheduler_test.go**: Updated test calls
- **internal/app/wiring_automation.go**: Updated `cron.New()` caller
- **internal/economy/escrow/store.go**: Interface + memoryStore impl
- **internal/economy/escrow/ent_store.go**: EntStore impl
- **internal/economy/escrow/hub/dangling_detector.go**: Uses filtered query
- **internal/app/convert.go**: New shared converter
- **internal/app/bridge_team_budget.go**: Uses shared converter
- **internal/app/bridge_team_escrow.go**: Uses shared converter
- **internal/p2p/gitbundle/bundle.go**: `runGit()` helper, refactored methods
