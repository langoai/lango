## 1. Wrapper Guard Coverage

- [x] 1.1 Add regression coverage for missing `command` on `exec` and `exec_bg`.
- [x] 1.2 Add regression coverage for missing `id` on `exec_status` and `exec_stop`.

## 2. Downstream Sync

- [x] 2.1 Update exec prompt wording for the required-input contract.
- [x] 2.2 Update operator-facing docs for the same contract.
- [x] 2.3 Update exec and production-readiness specs for wrapper guard coverage.

## 3. Verification

- [x] 3.1 Run `go test ./internal/tools/exec -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
