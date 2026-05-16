## 1. UX Fix

- [x] 1.1 Render an explicit no-categories message in the left Tools category pane.
- [x] 1.2 Add regression coverage for a configured catalog with zero categories.

## 2. Docs And Spec Sync

- [x] 2.1 Update the Tools page spec to require the explicit no-categories message.
- [x] 2.2 Update cockpit feature docs to describe the same empty-category surface.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
