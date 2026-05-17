## 1. Tests

- [x] 1.1 Add shared prompt helper tests for explicit input-stream interactive guards.
- [x] 1.2 Add failing onboard/settings tests proving the guard receives `cmd.InOrStdin()`.
- [x] 1.3 Add failing security/secrets tests proving interactive guards receive command input streams.

## 2. Implementation

- [x] 2.1 Add the shared command-input-aware interactive guard helper.
- [x] 2.2 Route onboard/settings interactive guards through command input.
- [x] 2.3 Route security passphrase/recovery/keyring/secrets interactive guards through command input.
- [x] 2.4 Preserve existing guidance messages, prompt output streams, and command behavior.

## 3. Verification

- [x] 3.1 Validate the OpenSpec change in strict mode.
- [x] 3.2 Run focused prompt/security/onboard/settings tests.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync and archive the OpenSpec change.
- [x] 3.5 Commit this scoped unit separately.
