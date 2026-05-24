## 1. Telegram Body Read Error Contract

- [ ] 1.1 Add a regression for Telegram download body-read failures.
- [ ] 1.2 Require the underlying read error cause to remain visible.

## 2. Verification

- [ ] 2.1 Run `go test ./internal/channels/telegram -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [ ] 3.1 Update `production-readiness` spec coverage for Telegram body-read failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
