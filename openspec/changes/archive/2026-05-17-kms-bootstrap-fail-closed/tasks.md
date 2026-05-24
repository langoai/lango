## 1. Tests

- [x] 1.1 Add failing bootstrap test for KMS provider initialization failure with fallback disabled.
- [x] 1.2 Add failing bootstrap test for KMS unwrap failure with fallback disabled.
- [x] 1.3 Add regression test showing fallback-enabled behavior still reaches passphrase acquisition.

## 2. Implementation

- [x] 2.1 Add a test seam for the bootstrap KMS provider factory.
- [x] 2.2 Make KMS provider initialization failure honor `FallbackToLocal`.
- [x] 2.3 Make KMS unwrap failure honor `FallbackToLocal`.
- [x] 2.4 Preserve existing warning text and passphrase fallback when fallback is enabled.

## 3. Specs and Verification

- [x] 3.1 Sync main `cloud-kms` and `bootstrap-lifecycle` specs.
- [x] 3.2 Validate the OpenSpec change in strict mode.
- [x] 3.3 Run focused bootstrap tests.
- [x] 3.4 Update public KMS docs for bootstrap fallback env override.
- [x] 3.5 Run `go build ./...` and `go test ./...`.
- [x] 3.6 Run subagent-driven review.
- [x] 3.7 Archive the OpenSpec change.
- [x] 3.8 Commit this scoped unit separately.
