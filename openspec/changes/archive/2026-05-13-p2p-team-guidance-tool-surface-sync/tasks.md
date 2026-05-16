## 1. CLI Guidance Fix

- [x] 1.1 Replace vague `p2p team` guidance strings with concrete `team_*` tool references.
- [x] 1.2 Add CLI regressions covering the updated team guidance strings.

## 2. Docs And Spec Sync

- [x] 2.1 Update the public P2P CLI docs to the same concrete team-tool guidance.
- [x] 2.2 Update the CLI P2P management spec to require concrete team-tool references instead of vague runtime wording.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/p2p -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
