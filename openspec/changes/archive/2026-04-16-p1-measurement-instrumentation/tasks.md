## 1. Team task delegation metrics (Unit 4 prep)

- [x] 1.1 Create `internal/app/bridge_team_metrics.go` — EventBus subscriber for `TeamTaskDelegatedEvent` and `TeamTaskCompletedEvent`, log structured metrics (worker count, success/fail, avg duration, dupe ratio estimate)
- [x] 1.2 Wire the subscriber in `modules.go:835` alongside other team bridges
- [x] 1.3 Unit test not needed — logging-only bridge, verified via build + existing integration tests

## 2. Child session lifecycle logging (Unit 5 prep)

- [x] 2.1 In `internal/app/wiring.go:703` childHook, added `logger().Infow("child session lifecycle", ...)` for all event types with childKey, parentKey, agentName
- [x] 2.2 Existing tests pass — no behavioral changes

## 3. Workspace cleanup error logging (Unit 9 prep)

- [x] 3.1 In `internal/runledger/workspace.go:106-110`, replaced `_ = m.RemoveWorktree(path)` and `_ = m.DeleteBranch(branch)` with `logging.App().Warnw(...)` on error
- [x] 3.2 Existing workspace tests pass

## 4. Verification

- [x] 4.1 `go build ./...` passes
- [x] 4.2 `go test ./...` passes — zero FAIL
