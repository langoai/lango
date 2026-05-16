## 1. Deterministic Clock-Sensitive Verification

- [x] 1.1 Update the proposal module learning-suggestion regression to use execution-relative timestamps so proposal visibility does not depend on the current calendar date.
- [x] 1.2 Update Mission Control loop regressions to inject the projector clock anywhere assertions depend on cron freshness or inquiry follow-up age windows.
- [x] 1.3 Align the integrated Mission Control loop expectation with the current design by asserting both the inquiry loop and its follow-up loop when the inquiry age crosses the follow-up threshold.

## 2. Parallel Executor Observability

- [x] 2.1 Ensure `ParallelReadOnlyExecutor` reports a positive duration for every completed eligible invocation, including handler error paths.
- [x] 2.2 Re-run the focused `internal/streamx` regression to confirm the failing-handler duration contract stays green.

## 3. Repository Verification

- [x] 3.1 Run `gofmt -w` on all modified Go files.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
