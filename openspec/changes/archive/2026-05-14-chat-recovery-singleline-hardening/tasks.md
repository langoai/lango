## 1. Recovery Row Hardening

- [x] 1.1 Normalize recovery `causeClass` text to a single line before rendering.
- [x] 1.2 Add regressions for multiline recovery metadata.

## 2. Spec Sync

- [x] 2.1 Record the recovery-row single-line contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to reflect the runtime behavior precisely.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-recovery-singleline-hardening --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
