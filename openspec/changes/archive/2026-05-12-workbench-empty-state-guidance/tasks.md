## 1. Workbench Empty-State Guidance

- [x] 1.1 Add regressions for incomplete and ready-profile workbench empty states.
- [x] 1.2 Detect incomplete profile setup in the Mission Control workbench surface.
- [x] 1.3 Render setup recovery guidance only for incomplete workbench profiles.

## 2. Downstream Documentation

- [x] 2.1 Update README workbench documentation to mention the setup recovery path.
- [x] 2.2 Update CLI/TUI docs to mention the incomplete-profile guidance in the workbench empty state.

## 3. Verification

- [x] 3.1 Run `gofmt -w` on modified Go files.
- [x] 3.2 Run `go test ./internal/cli/workbench ./internal/cli/cockpit/pages ./cmd/lango -count=1`.
- [ ] 3.3 Run `go build ./...`.
- [ ] 3.4 Run `go test ./...`.
- [ ] 3.5 Validate and archive the OpenSpec change.
