## 1. Telegram Download Contract

- [x] 1.1 Add a regression that inspects the outgoing Telegram file-download request.
- [x] 1.2 Assert HTTP GET semantics and the 30-second timeout-derived deadline.

## 2. Verification

- [x] 2.1 Run `go test ./internal/channels/telegram -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.
- [ ] 2.4 Validate and archive the OpenSpec change.
