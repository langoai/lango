## 1. UX Fix

- [x] 1.1 Render `Enter` as `seed starter` when the empty workbench will stage the default starter prompt.
- [x] 1.2 Render `Enter` as `run starter` when the empty workbench already has a staged starter prompt.
- [x] 1.3 Add regression coverage for both help states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require the context-sensitive empty-workbench `Enter` help label.
- [x] 2.2 Update cockpit feature docs to describe the same starter flow.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
