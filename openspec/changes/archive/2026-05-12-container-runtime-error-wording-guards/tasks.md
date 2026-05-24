## 1. Error Wording Regressions

- [x] 1.1 Add regression coverage for explicit Docker unavailable wording.
- [x] 1.2 Add regression coverage for explicit gVisor unavailable wording.
- [x] 1.3 Add regression coverage for fail-closed require-container unavailable wording.

## 2. Verification

- [x] 2.1 Run `go test ./internal/sandbox -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update production-readiness spec for actionable sandbox runtime unavailable errors.
- [ ] 3.2 Validate and archive the OpenSpec change.
