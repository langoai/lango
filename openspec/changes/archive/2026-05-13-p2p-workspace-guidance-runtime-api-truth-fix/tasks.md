## 1. CLI Guidance Fix

- [x] 1.1 Replace stale "runtime API" wording in `lango p2p workspace create/list/join/leave`.
- [x] 1.2 Add CLI regressions covering the new workspace guidance strings.

## 2. Spec Sync

- [x] 2.1 Update the CLI P2P management spec to forbid the stale runtime-API wording for workspace guidance commands.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/p2p -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
