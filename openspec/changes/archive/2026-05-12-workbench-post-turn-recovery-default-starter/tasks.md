## 1. Recovery Default Starter

- [x] 1.1 Add a recovery-specific default starter helper for failed completed-turn workbench states.
- [x] 1.2 Update the workbench page to seed that recovery default on failed completed turns.
- [x] 1.3 Add regressions for failed-turn recovery default behavior.
- [x] 1.4 Update public docs for the recovery-oriented `Enter` behavior.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/workbenchstart ./internal/cli/workbench -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `mission-workbench-tui` and `downstream-docs-sync` coverage for the recovery-oriented `Enter` behavior.
- [ ] 3.2 Validate and archive the OpenSpec change.
