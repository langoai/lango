## 1. Runtime Hardening

- [x] 1.1 Disable unregistered optional cockpit pages in the sidebar.
- [x] 1.2 Re-enable sidebar items when those pages are registered.
- [x] 1.3 Add regression coverage for disabled and enabled states.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/cockpit ./internal/cli/cockpit/sidebar -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.
- [x] 2.4 Validate and archive the OpenSpec change.
