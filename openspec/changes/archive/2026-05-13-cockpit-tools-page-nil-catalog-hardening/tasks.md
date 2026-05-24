## 1. Runtime Hardening

- [x] 1.1 Make ToolsPage nil-safe when ToolCatalog is unavailable.
- [x] 1.2 Register the Tools page even when the app has no tool catalog.
- [x] 1.3 Add regression coverage for the nil-catalog empty state.

## 2. Spec Sync

- [x] 2.1 Update the cockpit tools-page spec to require the degraded empty-state behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
