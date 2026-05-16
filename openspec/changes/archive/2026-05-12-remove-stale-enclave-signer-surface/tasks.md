## 1. Stale Signer Surface Removal

- [x] 1.1 Remove `enclave` from signer-provider validation and settings options.
- [x] 1.2 Treat KMS providers as valid in `lango doctor` security checks.
- [x] 1.3 Add regressions for invalid `enclave` config and valid KMS doctor checks.

## 2. Verification

- [x] 2.1 Run `go test ./internal/config ./internal/cli/doctor/checks -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update production-readiness spec for the supported signer-provider set.
- [ ] 3.2 Validate and archive the OpenSpec change.
