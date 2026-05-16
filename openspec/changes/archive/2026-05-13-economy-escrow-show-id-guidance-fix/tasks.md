## 1. CLI Guidance Fix

- [x] 1.1 Replace vague live-data guidance in `lango economy escrow show --id`.
- [x] 1.2 Add a CLI regression covering the new `escrow_status` guidance.

## 2. Docs And Spec Sync

- [x] 2.1 Update the economy CLI docs for `--id` guidance.
- [x] 2.2 Update the on-chain escrow spec to the same live-status contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/economy -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
