## 1. System Block Hardening

- [x] 1.1 Make system transcript blocks width-aware.
- [x] 1.2 Add regressions for narrow-width system rendering.

## 2. Spec Sync

- [x] 2.1 Record the system-block width-safety contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-system-block-width-safety --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
