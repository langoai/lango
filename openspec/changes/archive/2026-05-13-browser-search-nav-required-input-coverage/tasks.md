## 1. Browser Guard Coverage

- [x] 1.1 Add regression coverage for missing `url` on `browser_navigate`.
- [x] 1.2 Add regression coverage for missing `query` on `browser_search`.

## 2. Downstream Sync

- [x] 2.1 Update browser prompt wording for the search/navigation input contract.
- [x] 2.2 Update public multi-agent docs for the same contract.
- [x] 2.3 Update browser and production-readiness specs for the required-input coverage.

## 3. Verification

- [x] 3.1 Run `go test ./internal/tools/browser -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
