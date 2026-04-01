## 1. Core Types and Shell Unwrap

- [x] 1.1 Create `internal/tools/exec/policy.go` with `Verdict`, `ReasonCode`, `PolicyDecision`, `EventPublisher`, `PolicyEvaluator` types and `NewPolicyEvaluator` constructor
- [x] 1.2 Create `internal/tools/exec/unwrap.go` with `unwrapShellWrapper`, `isShellWrapper`, `stripQuotes` functions
- [x] 1.3 Create `internal/tools/exec/opaque.go` with `detectOpaquePattern` function
- [x] 1.4 Create `internal/tools/exec/unwrap_test.go` with table-driven tests for shell wrapper unwrap
- [x] 1.5 Create `internal/tools/exec/opaque_test.go` with table-driven tests for opaque pattern detection

## 2. PolicyEvaluator Logic

- [x] 2.1 Implement `PolicyEvaluator.Evaluate` method (unwrap → classify → guard → opaque → verdict)
- [x] 2.2 Implement `PolicyEvaluator.publishAndLog` method (structured logging + event publishing)
- [x] 2.3 Create `internal/tools/exec/policy_test.go` with table-driven tests for Evaluate (block/observe/allow scenarios)

## 3. Middleware and Event

- [x] 3.1 Create `internal/tools/exec/middleware.go` with `WithPolicy` middleware function
- [x] 3.2 Add `EventPolicyDecision` constant and `PolicyDecisionEvent` struct to `internal/eventbus/events.go`
- [x] 3.3 Create `internal/tools/exec/middleware_test.go` (block→next not called, observe→next called, passthrough for non-exec tools, nil bus→no event)

## 4. Wiring and Integration

- [x] 4.1 Add `classifyLangoExec` to `internal/app/tools.go` returning `(string, execpkg.ReasonCode)`, refactor `blockLangoExec` to delegate
- [x] 4.2 Add PolicyEvaluator creation and `WithPolicy` middleware application to `internal/app/app.go` Phase B (after WithApproval)
- [x] 4.3 Run `go build ./...` and fix any compilation errors
- [x] 4.4 Run `go test ./...` and fix any test failures

## 5. Specs and Downstream

- [x] 5.1 Update `openspec/specs/exec-command-guard/spec.md` with shell unwrap and observe verdict requirements
- [x] 5.2 Update `openspec/specs/tool-exec/spec.md` with middleware execution order and PolicyEvaluator integration
- [x] 5.3 Update `prompts/TOOL_USAGE.md` with shell wrapper bypass enforcement notice
- [x] 5.4 Update `prompts/SAFETY.md` with shell wrapper safety guidance
