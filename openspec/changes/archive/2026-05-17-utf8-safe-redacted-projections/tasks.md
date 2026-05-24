## 1. Tests First

- [x] 1.1 Add failing `internal/security` tests for redaction behavior and UTF-8-safe projection truncation.

## 2. Implementation

- [x] 2.1 Update `RedactedProjection` truncation to preserve a valid UTF-8 prefix within the existing byte limit.

## 3. Downstream Artifacts

- [x] 3.1 Update payload-protection docs/specs to state that redacted projection truncation is UTF-8 safe.

## 4. Verification

- [x] 4.1 Run focused `internal/security` tests.
- [x] 4.2 Run `openspec validate utf8-safe-redacted-projections --strict`.
- [x] 4.3 Run `go build ./...` and `go test ./...`.
- [x] 4.4 Sync/archive the OpenSpec change after verification.
