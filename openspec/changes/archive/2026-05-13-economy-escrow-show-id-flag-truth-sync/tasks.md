## 1. CLI Help Sync

- [x] 1.1 Replace the stale `future use` wording for `lango economy escrow show --id`.
- [x] 1.2 Add a CLI regression covering the updated flag description.

## 2. Spec Sync

- [x] 2.1 Update the on-chain escrow spec to require truthful `--id` help wording.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/economy -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
