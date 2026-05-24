## 1. Invalid Recipient Hex Error Contract

- [x] 1.1 Add invalid-hex recipient coverage alongside the existing format-invalid cases.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` spec coverage for invalid recipient hex-cause preservation.
- [ ] 3.2 Validate and archive the OpenSpec change.
