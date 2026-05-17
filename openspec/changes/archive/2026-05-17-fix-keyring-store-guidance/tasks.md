## 1. Tests

- [x] 1.1 Add regression coverage proving stale keyring update warnings point to `lango security keyring store`.
- [x] 1.2 Prove the warning no longer references nonexistent `lango security keyring set`.

## 2. Implementation

- [x] 2.1 Centralize the stale keyring warning text shared by passphrase-changing commands.
- [x] 2.2 Replace the stale `keyring set` guidance with `keyring store`.
- [x] 2.3 Preserve existing command stderr routing for the warning.

## 3. Verification

- [x] 3.1 Validate the OpenSpec change in strict mode.
- [x] 3.2 Run focused security CLI tests.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync and archive the OpenSpec change.
- [x] 3.5 Commit this scoped unit separately.
