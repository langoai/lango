## 1. Wrapper Param Guards

- [x] 1.1 Add settlement progression wrapper coverage for missing `outcome`.
- [x] 1.2 Add escrow adjudication wrapper coverage for missing `outcome`.

## 2. Verification

- [x] 2.1 Run `go test ./internal/app -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `meta-tools` and `production-readiness` coverage for wrapper-level missing-`outcome` errors.
- [ ] 3.2 Validate and archive the OpenSpec change.
