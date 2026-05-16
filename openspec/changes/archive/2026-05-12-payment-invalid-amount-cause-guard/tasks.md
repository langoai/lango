## 1. Invalid Amount Error Contract

- [x] 1.1 Tighten the invalid amount regression to require the underlying parsing cause.
- [x] 1.2 Cover the too-many-decimals parse path.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` spec coverage for invalid amount error shape.
- [ ] 3.2 Validate and archive the OpenSpec change.
