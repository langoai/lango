## Why

The stdin-pipe passphrase path currently reports a raw `read stdin: EOF` failure when the pipe is empty. That is technically correct but not very helpful: the real problem is simply that no passphrase was supplied.

## What Changes

- Treat empty EOF on the stdin-pipe passphrase path as an empty-input passphrase error
- Add regressions for empty EOF and no-trailing-newline success
- Document the clarified stdin-pipe behavior in the passphrase spec and security docs

## Capabilities

### New Capabilities

### Modified Capabilities
- `passphrase-acquisition`: empty stdin EOF is classified as empty passphrase input

## Impact

- Affected code: `internal/security/passphrase/stdin.go`, `internal/security/passphrase/stdin_test.go`
- Affected docs: `docs/security/encryption.md`
- Affected specs: `openspec/specs/passphrase-acquisition/spec.md`
- No feature expansion; this is non-interactive UX and error-clarity hardening
