## 1. Shared Prompt Helpers

- [x] 1.1 Add failing tests for explicit-output passphrase confirmation.
- [x] 1.2 Implement `PassphraseConfirmIO` using existing hidden-input semantics.
- [x] 1.3 Verify prompt helper tests pass.

## 2. Security Command Routing

- [x] 2.1 Add failing security command tests for command-output passphrase prompt capture.
- [x] 2.2 Route keyring store, change-passphrase, migrate-passphrase, recovery setup, and recovery restore prompt text through command output streams.
- [x] 2.3 Preserve command error stream routing for keyfile/keyring notices and warnings.

## 3. Verification And OpenSpec

- [x] 3.1 Validate the OpenSpec change in strict mode.
- [x] 3.2 Run focused prompt and security tests.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync and archive the OpenSpec change after implementation.
- [x] 3.5 Commit this scoped unit separately.
