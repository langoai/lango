## 1. Payment History Contract

- [x] 1.1 Add a regression that verifies payment history is returned in descending `created_at` order.
- [x] 1.2 Assert that `History(limit=1)` returns the newest record.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.
- [ ] 2.4 Validate and archive the OpenSpec change.
