## 1. Tests First

- [x] 1.1 Add failing EIP-3009 tests for nonce reader failure.
- [x] 1.2 Add/update caller tests to prove paid invocation does not sign or invoke remotely when authorization creation fails.

## 2. Implementation

- [x] 2.1 Make `eip3009.NewUnsigned` return an error on nonce generation failure.
- [x] 2.2 Update paid invocation caller to wrap and return authorization creation errors before signing.
- [x] 2.3 Preserve existing successful EIP-3009 signing and calldata behavior.

## 3. Verification

- [x] 3.1 Run focused EIP-3009 and P2P paid invocation tests.
- [x] 3.2 Run `openspec validate eip3009-nonce-error-handling --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
