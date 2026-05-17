## 1. Tests

- [x] 1.1 Add failing tests for existing-passphrase interactive acquisition prompt routing.
- [x] 1.2 Add failing tests for first-run passphrase confirmation prompt routing.

## 2. Implementation

- [x] 2.1 Route interactive acquisition prompts through the `acquireWithIO` writer.
- [x] 2.2 Preserve hidden terminal input and acquisition priority behavior.

## 3. Verification

- [x] 3.1 Validate the OpenSpec change in strict mode.
- [x] 3.2 Run focused passphrase acquisition tests.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync and archive the OpenSpec change.
- [x] 3.5 Commit this scoped unit separately.
