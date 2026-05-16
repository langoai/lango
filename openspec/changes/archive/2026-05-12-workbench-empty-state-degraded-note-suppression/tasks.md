## 1. Empty-State Degraded-Note Suppression

- [x] 1.1 Suppress degraded-note rendering on the empty standalone workbench surface.
- [x] 1.2 Add regressions for workbench-vs-cockpit degraded-note behavior.
- [x] 1.3 Update public docs for the calmer empty-state warning posture.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `mission-workbench-tui` and `downstream-docs-sync` coverage for empty-state degraded-note suppression.
- [ ] 3.2 Validate and archive the OpenSpec change.
