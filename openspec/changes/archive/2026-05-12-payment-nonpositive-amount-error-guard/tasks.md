## 1. Non-Positive Amount Error Contract

- [x] 1.1 Tighten the zero/negative amount regression to keep the explicit business-rule error path.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` spec coverage for non-positive amount error shape.
- [ ] 3.2 Validate and archive the OpenSpec change.
