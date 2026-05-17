## 1. Tests First

- [x] 1.1 Add failing table-output test that bootstrap errors appear as a dedicated doctor result with root-cause details.
- [x] 1.2 Add failing JSON-output test that bootstrap errors appear as a dedicated doctor result with root-cause details.

## 2. Implementation

- [x] 2.1 Add a failing bootstrap diagnostic result when doctor bootstrap fails.
- [x] 2.2 Keep remaining checks running with nil config after bootstrap failure.
- [x] 2.3 Avoid closing nil bootstrap results.

## 3. Docs and Verification

- [x] 3.1 Update public doctor CLI docs for bootstrap failure reporting.
- [x] 3.2 Run focused doctor tests and OpenSpec validation.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Run subagent review.
- [x] 3.5 Sync and archive the OpenSpec change.
