## 1. Tests First

- [x] 1.1 Add failing test that mutating the `Allocate` return value does not mutate stored budget state.
- [x] 1.2 Add failing test that mutating `Get` and `List` results does not mutate stored budget state.
- [x] 1.3 Add failing test that mutating a budget after `Update` does not mutate stored budget state.

## 2. Implementation

- [x] 2.1 Add deep-copy support for `TaskBudget` snapshots.
- [x] 2.2 Return detached snapshots from `Allocate`, `Get`, and `List`.
- [x] 2.3 Store detached snapshots in `Update`.
- [x] 2.4 Preserve existing budget engine behavior.

## 3. Verification

- [x] 3.1 Run focused budget tests and OpenSpec validation.
- [x] 3.2 Run `go build ./...` and `go test ./...`.
- [x] 3.3 Run subagent review.
- [x] 3.4 Sync and archive the OpenSpec change.
