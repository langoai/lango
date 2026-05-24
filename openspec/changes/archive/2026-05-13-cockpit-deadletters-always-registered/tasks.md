## 1. Runtime Wiring

- [x] 1.1 Always register the Dead Letters page in cockpit startup.
- [x] 1.2 Keep a degraded empty/unavailable state when the dead-letter bridge is absent.
- [x] 1.3 Add regressions for wiring and nil-list activation behavior.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs/specs to describe Dead Letters as always registered with degraded unavailable messaging.

## 3. Verification

- [x] 3.1 Run `go test ./cmd/lango ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
