## 1. UX Fix

- [x] 1.1 Render explicit unavailable messaging in the context panel when the metrics collector is absent.
- [x] 1.2 Add regression coverage for the nil-collector context panel view.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to describe the context panel unavailable-state behavior.
- [x] 2.2 Update the cockpit context-panel spec to require unavailable messaging.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
