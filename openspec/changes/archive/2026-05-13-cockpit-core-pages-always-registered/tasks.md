## 1. Runtime Wiring

- [x] 1.1 Always register the Status page in cockpit startup.
- [x] 1.2 Always register the Settings page in cockpit startup.
- [x] 1.3 Add a regression that locks the wiring contract.

## 2. Spec Sync

- [x] 2.1 Update cockpit page specs to require always-registered Status and Settings surfaces.

## 3. Verification

- [x] 3.1 Run `go test ./cmd/lango ./internal/cli/cockpit -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
