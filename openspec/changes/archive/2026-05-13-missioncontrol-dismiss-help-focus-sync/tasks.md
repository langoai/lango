## 1. UX Fix

- [x] 1.1 Show `d` only when a proposed mission is selected and the missions lane is focused.
- [x] 1.2 Add regression coverage for focused vs non-focused dismiss help states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require the focus-sensitive dismiss help binding.
- [x] 2.2 Update cockpit feature docs to describe the same focus condition.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
