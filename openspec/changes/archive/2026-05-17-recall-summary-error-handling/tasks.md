## 1. Tests First

- [x] 1.1 Add failing focused tests for recall adapter summary-load failure.

## 2. Implementation

- [x] 2.1 Return an actionable error from `RecallRecent` when `GetSummary` fails for a retained hit.
- [x] 2.2 Preserve current-session exclusion, rank-floor filtering, and top-N behavior.

## 3. Verification

- [x] 3.1 Run focused `internal/app` recall tests.
- [x] 3.2 Run `openspec validate recall-summary-error-handling --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
