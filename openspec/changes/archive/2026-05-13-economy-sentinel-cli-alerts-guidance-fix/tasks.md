## 1. CLI Guidance Fix

- [x] 1.1 Replace the nonexistent `lango economy escrow sentinel alerts` guidance in sentinel status output.
- [x] 1.2 Add a CLI regression covering the new sentinel guidance.

## 2. Docs And Spec Sync

- [x] 2.1 Update the economy CLI docs to the same sentinel guidance contract.
- [x] 2.2 Update the sentinel spec so CLI docs do not imply a nonexistent alerts subcommand.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/economy -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
