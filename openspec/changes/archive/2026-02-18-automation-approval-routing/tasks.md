## 1. Context Helpers

- [x] 1.1 Create `internal/approval/context.go` with `WithApprovalTarget` and `ApprovalTargetFromContext` functions

## 2. Approval Routing Override

- [x] 2.1 Modify `wrapWithApproval` in `internal/app/tools.go` to check approval target before using session key

## 3. Automation System Integration

- [x] 3.1 Inject approval target from `job.DeliverTo[0]` in `internal/cron/executor.go` (only when contains `:`)
- [x] 3.2 Inject approval target from `task.OriginSession` or `task.OriginChannel` in `internal/background/manager.go`

## 4. Verification

- [x] 4.1 Run `go build ./...` and `go test ./...` to verify no regressions
