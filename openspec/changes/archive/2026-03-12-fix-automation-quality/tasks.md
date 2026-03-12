## 1. Import Ordering

- [x] 1.1 Move `"time"` import into stdlib group in `internal/app/tools_automation.go`

## 2. Scheduler Stop Idempotency

- [x] 2.1 Add `stopOnce sync.Once` field to `Scheduler` struct
- [x] 2.2 Wrap `Stop()` body in `s.stopOnce.Do(func() { ... })`

## 3. Remove Redundant unregisterJob

- [x] 3.1 Remove `s.unregisterJob(job.ID)` call from `disableOneTimeJob`
- [x] 3.2 Add comment clarifying that unregisterJob is handled by sync.Once in registerJob

## 4. Upsert Returns Persisted Job

- [x] 4.1 Change `Store.Upsert` interface signature from `(bool, error)` to `(*Job, bool, error)`
- [x] 4.2 Update `EntStore.Upsert` to return `*Job` (read-back on create path, return in-memory on update path)
- [x] 4.3 Update `Scheduler.AddJob` to use returned `*Job` directly, remove `GetByName` call
- [x] 4.4 Update `mockStore.Upsert` in `scheduler_test.go` to match new signature
- [x] 4.5 Update `MockCronStore.Upsert` in `testutil/mock_cron.go` to match new signature

## 5. Remove Unused Test Mock Field

- [x] 5.1 Remove `delay time.Duration` field from `mockAgentRunner` struct
- [x] 5.2 Remove `time.Sleep(m.delay)` logic from `mockAgentRunner.Run()`

## 6. Verification

- [x] 6.1 Run `go build ./...` — passes
- [x] 6.2 Run `go test ./internal/cron/... ./internal/background/... ./internal/workflow/...` — all pass
- [x] 6.3 Run `go vet ./...` — no issues
