## 1. Completed-Turn Failure Wording

- [x] 1.1 Update completed-turn empty body wording to call out failed turns that need attention.
- [x] 1.2 Add regressions for the failure-aware wording path.
- [x] 1.3 Update public docs for the failure-aware completed-turn wording.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/workbench -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `mission-workbench-tui` and `downstream-docs-sync` coverage for the failure-aware completed-turn wording.
- [ ] 3.2 Validate and archive the OpenSpec change.
