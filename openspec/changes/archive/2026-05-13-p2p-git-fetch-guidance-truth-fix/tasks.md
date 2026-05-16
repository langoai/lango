## 1. CLI Guidance Fix

- [x] 1.1 Replace nonexistent `p2p_git_fetch` guidance in `lango p2p git fetch`.
- [x] 1.2 Add a CLI regression covering the fetch guidance output.

## 2. Spec Sync

- [x] 2.1 Update the P2P CLI management spec to the same fetch guidance contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/p2p -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
