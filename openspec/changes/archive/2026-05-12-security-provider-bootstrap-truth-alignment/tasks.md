## 1. Security Wiring Error Contracts

- [x] 1.1 Tighten the local security-provider bootstrap error regression.
- [x] 1.2 Tighten the KMS security-provider build-tag error regression.

## 2. Downstream Truth Alignment

- [x] 2.1 Update security provider docs to list only the current supported values.
- [x] 2.2 Document bootstrap-backed wiring requirements for local providers and build-tag requirements for KMS providers.

## 3. Verification

- [x] 3.1 Run `go test ./internal/app -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.

## 4. Spec Sync

- [x] 4.1 Update production-readiness and downstream-docs-sync specs for provider/bootstrap truth alignment.
- [ ] 4.2 Validate and archive the OpenSpec change.
