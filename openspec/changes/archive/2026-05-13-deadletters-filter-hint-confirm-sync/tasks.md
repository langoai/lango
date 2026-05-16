## 1. UX Fix

- [x] 1.1 Render a confirm-state filter hint when retry confirmation is pending.
- [x] 1.2 Add regression coverage for the confirm-state hint text.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require the confirm-state hint wording.
- [x] 2.2 Update cockpit feature docs to describe the confirm-state `Enter`/`Esc` hint.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
