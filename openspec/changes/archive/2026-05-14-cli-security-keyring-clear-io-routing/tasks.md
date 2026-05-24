## 1. Command Streams

- [x] 1.1 Route `lango security keyring clear` prompt and result output through the Cobra command writer.
- [x] 1.2 Route `lango security keyring clear` warnings through the Cobra error writer.
- [x] 1.3 Read `lango security keyring clear` confirmation input through the Cobra command input reader.
- [x] 1.4 Add command-level capture tests for abort, confirm, and `--force` flows.

## 2. Spec Sync

- [x] 2.1 Record the keyring clear I/O routing contract in `keyring-security-tiering`.
- [x] 2.2 Update downstream `docs/cli/security.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/security -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-security-keyring-clear-io-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
