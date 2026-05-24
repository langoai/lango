## 1. Tests

- [x] 1.1 Add a failing settings form test for KMS fallback copy covering signing, encryption, unwrap, and bootstrap env override.

## 2. Implementation

- [x] 2.1 Update the Security KMS fallback field description to match runtime behavior.
- [x] 2.2 Preserve the existing KMS form field set and save behavior.

## 3. Specs and Verification

- [x] 3.1 Sync the `settings-security-advanced` main spec.
- [x] 3.2 Validate the OpenSpec change in strict mode.
- [x] 3.3 Run focused settings tests.
- [x] 3.4 Run `go build ./...` and `go test ./...`.
- [x] 3.5 Run subagent-driven review.
- [x] 3.6 Archive the OpenSpec change.
- [x] 3.7 Commit this scoped unit separately.
