## 1. UX Fix

- [x] 1.1 Surface an explicit degraded note when Mission Control has no projector.
- [x] 1.2 Add regression coverage for the nil-projector degraded state.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to describe the nil-projector degraded behavior.
- [x] 2.2 Update cockpit page specs to require the degraded note.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
