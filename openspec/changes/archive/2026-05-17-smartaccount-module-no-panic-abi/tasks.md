## 1. Tests First

- [x] 1.1 Add a failing `internal/smartaccount/module` package test that rejects production `panic(` calls.
- [x] 1.2 Run the focused test and confirm it fails on the existing ABI type helper panic.

## 2. Implementation

- [x] 2.1 Replace panic-based ABI type construction with checked initialization that returns encoder errors.
- [x] 2.2 Preserve install/uninstall calldata selectors, ABI layout, and public function signatures.

## 3. Verification

- [x] 3.1 Run focused smartaccount module tests.
- [x] 3.2 Run `openspec validate smartaccount-module-no-panic-abi --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
