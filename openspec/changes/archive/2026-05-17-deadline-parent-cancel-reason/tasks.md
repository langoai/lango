## 1. Tests First

- [x] 1.1 Add a failing `internal/deadline` test for parent cancellation reason classification.
- [x] 1.2 Add a failing `internal/app` alias test for the same parent cancellation behavior.

## 2. Implementation

- [x] 2.1 Update `internal/deadline.New` to observe parent cancellation and classify it as `ReasonCancelled`.
- [x] 2.2 Keep idle, max-timeout, extend, and stop behavior unchanged.

## 3. Verification

- [x] 3.1 Run focused deadline/app tests.
- [x] 3.2 Run `openspec validate deadline-parent-cancel-reason --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
