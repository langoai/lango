## 1. UX Fix

- [x] 1.1 Fail closed when accepting a proposed mission without a mission service.
- [x] 1.2 Fail closed when dismissing a proposed mission without a proposal service.
- [x] 1.3 Add regressions for both silent-no-op edges.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to describe the fail-closed proposal action behavior.
- [x] 2.2 Update cockpit page specs to require explicit system messages for missing services.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
