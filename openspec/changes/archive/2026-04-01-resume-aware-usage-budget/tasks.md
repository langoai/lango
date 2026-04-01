## 1. BudgetPolicy Serialization

- [x] 1.1 Add `Serialize() map[string]string` method to BudgetPolicy in `internal/agentrt/budget.go`
- [x] 1.2 Add `Restore(state map[string]string)` method to BudgetPolicy in `internal/agentrt/budget.go`
- [x] 1.3 Add unit tests for Serialize and Restore in `internal/agentrt/budget_test.go`

## 2. Expose Budget from initAgentRuntime

- [x] 2.1 Change `initAgentRuntime` return type to `(turnrunner.Executor, *agentrt.BudgetPolicy)` in `internal/app/wiring_agentrt.go`
- [x] 2.2 Return `budget` in structured mode and `nil` in classic mode

## 3. Session Usage Wiring

- [x] 3.1 Create `internal/app/wiring_session_usage.go` with `budgetRestoringExecutor` struct implementing `turnrunner.Executor`
- [x] 3.2 Implement lazy restore logic: read session metadata on first call per session key using `sync.Map`
- [x] 3.3 Implement `wireSessionUsage` function that registers OnTurnComplete callback to persist budget + token state
- [x] 3.4 Add unit tests for `budgetRestoringExecutor` and `wireSessionUsage` in `internal/app/wiring_session_usage_test.go`

## 4. App Wiring Integration

- [x] 4.1 Update `app.go` to capture budget from `initAgentRuntime` return
- [x] 4.2 Wrap executor with `budgetRestoringExecutor` when budget is non-nil
- [x] 4.3 Call `wireSessionUsage` after TurnRunner creation

## 5. Verification

- [x] 5.1 Run `go build ./...` to verify compilation
- [x] 5.2 Run `go test ./internal/agentrt/... -v` to verify budget tests
- [x] 5.3 Run `go test ./internal/app/... -v` to verify wiring tests
