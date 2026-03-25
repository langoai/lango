## 1. Control Plane Fixes

- [x] 1.1 Add `adk.ChainOnEvent()` so wrappers can preserve existing `onEvent` handlers
- [x] 1.2 Update `CoordinatingExecutor` to use `ChainOnEvent()` instead of overwriting `onEvent`
- [x] 1.3 Move mutable delegation target and mirrored budget state to per-run/per-attempt state containers
- [x] 1.4 Ensure circuit breaker outcomes are recorded for the delegated specialist from the current attempt only

## 2. Trace Metrics Fixes

- [x] 2.1 Update `ComputeAgentMetrics` to attribute non-delegated turns from trace/event agent evidence
- [x] 2.2 Remove current-config-mode-based root attribution from `lango agent trace metrics`
- [x] 2.3 Add regression tests for non-delegated single-agent and multi-agent attribution

## 3. Verification

- [x] 3.1 Add regression tests for preserved `onEvent` chaining
- [x] 3.2 Add regression tests for correct breaker outcome attribution across delegation-return flows
- [x] 3.3 Run `go build ./...`
- [x] 3.4 Run `go test ./...`
