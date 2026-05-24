## 1. CLI Help Sync

- [x] 1.1 Replace vague top-level `p2p team` and `p2p workspace` help wording with concrete tool-surface references.
- [x] 1.2 Add CLI regressions covering the updated help output.

## 2. Docs And Spec Sync

- [x] 2.1 Update the public P2P CLI docs to use the same concrete top-level help wording.
- [x] 2.2 Update the CLI P2P management spec to require concrete tool-surface references in top-level help.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/p2p -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
