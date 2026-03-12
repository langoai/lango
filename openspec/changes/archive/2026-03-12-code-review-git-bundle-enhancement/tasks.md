## 1. Critical Fixes

- [x] 1.1 Separate ping context from git state context in `health_monitor.go:pingMember()`
- [x] 1.2 Guard event bus subscriptions with `sync.Once` in `health_monitor.go:Start()`
- [x] 1.3 Add `subsOnce sync.Once` field to `HealthMonitor` struct

## 2. Major Fixes

- [x] 2.1 Add `GetByToken(token string) (string, bool)` to `SessionStore` in `handshake/session.go`
- [x] 2.2 Update session validator closure in `app.go` to use `GetByToken()`
- [x] 2.3 Add slice capacity hints in `tools_team.go` for team_form and team_status handlers
- [x] 2.4 Create `SchedulerConfig` struct in `cron/scheduler.go`
- [x] 2.5 Refactor `cron.New()` to accept `SchedulerConfig` instead of 6 positional params
- [x] 2.6 Update `initCron()` caller in `wiring_automation.go`
- [x] 2.7 Update all `cron.New()` calls in `scheduler_test.go`
- [x] 2.8 Add `ListByStatusBefore(status, time.Time)` to escrow `Store` interface
- [x] 2.9 Implement `ListByStatusBefore` in `memoryStore`
- [x] 2.10 Implement `ListByStatusBefore` in `EntStore` using ent `CreatedAtLT` predicate
- [x] 2.11 Update `DanglingDetector.scan()` to use `ListByStatusBefore` with cutoff time

## 3. Minor Fixes

- [x] 3.1 Add 5s timeout context for DID lookup in `app.go`
- [x] 3.2 Create `internal/app/convert.go` with shared `floatToMicroUSDC()` function
- [x] 3.3 Replace `floatToBudgetAmount()` in `bridge_team_budget.go` with `floatToMicroUSDC()`
- [x] 3.4 Replace `floatToUSDC()` in `bridge_team_escrow.go` with `floatToMicroUSDC()`
- [x] 3.5 Extract `runGit()` helper in `gitbundle/bundle.go`
- [x] 3.6 Refactor `Diff()` to use `runGit()` helper
- [x] 3.7 Refactor `snapshotRefs()` to use `runGit()` helper

## 4. Verification

- [x] 4.1 `go build ./...` passes
- [x] 4.2 `go test ./internal/p2p/team/...` passes
- [x] 4.3 `go test ./internal/cron/...` passes
- [x] 4.4 `go test ./internal/economy/escrow/...` passes
- [x] 4.5 `go test ./internal/p2p/gitbundle/...` passes
- [x] 4.6 `go test ./internal/p2p/handshake/...` passes
- [x] 4.7 `go vet ./...` passes
