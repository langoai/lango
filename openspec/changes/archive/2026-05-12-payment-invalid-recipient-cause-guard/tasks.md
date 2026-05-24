## 1. Invalid Recipient Error Contract

- [x] 1.1 Tighten the invalid recipient regression to require the underlying validation cause.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` spec coverage for invalid recipient error shape.
- [ ] 3.2 Validate and archive the OpenSpec change.
