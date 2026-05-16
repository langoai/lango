## 1. UX Fix

- [x] 1.1 Expose `d` in Mission Control help when the selected row is a proposed mission.
- [x] 1.2 Add regression coverage for proposal and non-proposal help states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require the context-sensitive dismiss help binding.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
