## 1. Wrapper Guard

- [x] 1.1 Tighten `escrow_create` and `escrow_resolve` wrapper guards for required params.
- [x] 1.2 Add regression coverage for the missing-parameter wrapper paths.

## 2. Verification

- [x] 2.1 Run `go test ./internal/app -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `onchain-escrow` and `production-readiness` coverage for the wrapper guard contract.
- [x] 3.2 Validate and archive the OpenSpec change.
