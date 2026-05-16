## 1. Result Preview Dedup

- [x] 1.1 Strip the duplicated `Assistant reply:` prefix from successful completed-turn body previews.
- [x] 1.2 Add regressions for the deduplicated preview wording.
- [x] 1.3 Update public docs for the cleaner result preview.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/workbench -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `mission-workbench-tui` and `downstream-docs-sync` coverage for the deduplicated result preview.
- [ ] 3.2 Validate and archive the OpenSpec change.
