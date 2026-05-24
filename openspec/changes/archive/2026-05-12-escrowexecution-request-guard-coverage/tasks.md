## 1. Escrow Execution Request Guard Coverage

- [x] 1.1 Add a regression test for missing `transaction receipt id` in escrow execution entrypoints.

## 2. Verification

- [x] 2.1 Run `go test ./internal/escrowexecution -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for the escrow execution request-id guard.
- [ ] 3.2 Validate and archive the OpenSpec change.
