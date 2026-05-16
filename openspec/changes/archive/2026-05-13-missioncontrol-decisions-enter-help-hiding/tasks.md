## 1. UX Fix

- [x] 1.1 Hide the generic `Enter` help while Mission Control decisions focus is active and no proposal-accept path applies.
- [x] 1.2 Add regression coverage for decisions-focus help.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require the reduced decisions-lane help surface.
- [x] 2.2 Update cockpit feature docs to describe that `Enter` is not a decisions-lane approval key.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
