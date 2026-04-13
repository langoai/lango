## 1. Core Implementation

- [x] 1.1 Create `internal/streamx/agent_fanin.go` with `AgentStreamFanIn` struct, `NewAgentStreamFanIn`, `AddChild`, `MergedStream`
- [x] 1.2 Implement `wrapChild` to detect per-child completion/error and emit progress events
- [x] 1.3 Implement `emptyTagStream` for empty children case
- [x] 1.4 Implement `emitStarted`, `emitCompleted`, `emitFailed` with nil bus guard

## 2. Tests

- [x] 2.1 Create `internal/streamx/agent_fanin_test.go` with `testStringStream` helper
- [x] 2.2 Test: two children merged output contains all events tagged correctly
- [x] 2.3 Test: single child degenerate case
- [x] 2.4 Test: empty children returns empty stream
- [x] 2.5 Test: ProgressBus receives lifecycle events (started + completed)
- [x] 2.6 Test: one child error, others continue, ProgressFailed emitted
- [x] 2.7 Test: nil bus does not panic

## 3. Verification

- [x] 3.1 `go build ./...` passes
- [x] 3.2 `go test ./internal/streamx/...` all pass
- [x] 3.3 `go vet ./internal/streamx/...` clean
