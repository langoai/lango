## 1. Header Truth Alignment

- [x] 1.1 Mark incomplete profiles as `Setup required` in the shared Mission Control header summary.
- [x] 1.2 Add projector and workbench regressions covering both incomplete and ready profile summaries.

## 2. Downstream Sync

- [x] 2.1 Update README and CLI/TUI docs to mention the readiness-aware header summary.

## 3. Verification

- [x] 3.1 Run `gofmt -w` on modified Go files.
- [x] 3.2 Run focused cockpit/workbench tests for the new header behavior.
- [ ] 3.3 Run `go build ./...`.
- [ ] 3.4 Run `go test ./...`.
- [ ] 3.5 Validate and archive the OpenSpec change.
