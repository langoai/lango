## 1. Composer Guidance

- [x] 1.1 Add regressions asserting the workbench composer placeholder changes with profile readiness.
- [x] 1.2 Override the workbench empty-state composer placeholder with setup-first or starter-prompt guidance as appropriate.

## 2. Documentation

- [x] 2.1 Update README and CLI/TUI docs to mention the state-aware composer guidance.

## 3. Verification

- [x] 3.1 Run `gofmt -w` on modified Go files.
- [x] 3.2 Run `go test ./internal/cli/workbench ./internal/cli/cockpit/pages ./cmd/lango -count=1`.
- [ ] 3.3 Run `go build ./...`.
- [ ] 3.4 Run `go test ./...`.
- [ ] 3.5 Validate and archive the OpenSpec change.
