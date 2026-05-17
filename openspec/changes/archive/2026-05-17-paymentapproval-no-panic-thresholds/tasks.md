## 1. Tests First

- [x] 1.1 Add a failing paymentapproval package test that rejects production `panic(` calls.
- [x] 1.2 Run the focused test and confirm it fails on the existing `mustParseUSDC` panic.

## 2. Implementation

- [x] 2.1 Replace panic-based threshold parsing with deterministic non-panicking threshold construction.
- [x] 2.2 Keep existing upfront-payment decision behavior and threshold boundaries unchanged.

## 3. Verification

- [x] 3.1 Run focused paymentapproval tests.
- [x] 3.2 Run `openspec validate paymentapproval-no-panic-thresholds --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
