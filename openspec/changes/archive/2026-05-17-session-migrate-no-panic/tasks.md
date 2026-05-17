## 1. Tests First

- [x] 1.1 Add a failing session-store regression for a panicking `MigrateSecrets` re-encryption callback.
- [x] 1.2 Run the focused test and confirm it fails on the current re-panic behavior.

## 2. Implementation

- [x] 2.1 Convert recovered `MigrateSecrets` panics into returned errors after rollback.
- [x] 2.2 Preserve existing rollback behavior for ordinary errors.

## 3. Verification

- [x] 3.1 Run focused session package tests.
- [x] 3.2 Run `openspec validate session-migrate-no-panic --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
