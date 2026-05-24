## 1. Refined Post-Turn Default

- [x] 1.1 Add failing regressions for generic and repo-aware post-turn default starter behavior.
- [x] 1.2 Make the post-turn empty-state default choose a structure-oriented starter outside repo context and a next-change starter inside detected workspaces.
- [x] 1.3 Update public docs for the refined post-turn default behavior.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/workbenchstart ./internal/cli/workbench -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `mission-workbench-tui` and `downstream-docs-sync` coverage for the refined post-turn default behavior.
- [ ] 3.2 Validate and archive the OpenSpec change.
