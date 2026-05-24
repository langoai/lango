## 1. Ready-Profile Starter Guidance

- [x] 1.1 Add a regression asserting the ready-profile workbench empty state shows concrete starter prompts.
- [x] 1.2 Render starter prompts only when the workbench profile is ready and setup guidance is not needed.

## 2. Documentation And Spec Hygiene

- [x] 2.1 Update README and CLI/TUI docs to mention ready-profile starter prompts.
- [x] 2.2 Replace the placeholder purpose text in `openspec/specs/mission-workbench-tui/spec.md`.

## 3. Verification

- [x] 3.1 Run `gofmt -w` on modified Go files.
- [x] 3.2 Run `go test ./internal/cli/workbench ./internal/cli/cockpit/pages ./cmd/lango -count=1`.
- [ ] 3.3 Run `go build ./...`.
- [ ] 3.4 Run `go test ./...`.
- [ ] 3.5 Validate and archive the OpenSpec change.
