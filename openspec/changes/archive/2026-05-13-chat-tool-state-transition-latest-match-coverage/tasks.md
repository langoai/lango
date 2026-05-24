## 1. Coverage

- [x] 1.1 Add a regression proving tool state transitions update the latest matching running row.

## 2. Spec Sync

- [x] 2.1 Record the latest-match transition rule in the OpenSpec delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-tool-state-transition-latest-match-coverage --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
