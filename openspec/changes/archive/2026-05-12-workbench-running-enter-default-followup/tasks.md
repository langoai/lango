## 1. Running-State Enter Follow-Up

- [x] 1.1 Make `Enter` queue the default follow-up when the running-state composer is empty.
- [x] 1.2 Add a regression for the no-draft running-state Enter path.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/workbench -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `mission-workbench-tui` spec coverage for default running-state follow-up queuing.
- [ ] 3.2 Validate and archive the OpenSpec change.
