## 1. UX Fix

- [x] 1.1 Make ordinary Mission Control submits fail closed when mission service is absent.
- [x] 1.2 Preserve slash-command passthrough.
- [x] 1.3 Add regression coverage for the fail-closed submit path.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to describe the fail-closed behavior.
- [x] 2.2 Update cockpit page specs to require the fail-closed submit path.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
