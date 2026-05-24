## 1. Tests

- [x] 1.1 Add failing command tests proving representative P2P CLI paths pass `cmd.Context()` into their loader seams.
- [x] 1.2 Add failing test coverage proving `initP2PDeps` passes caller context into `Node.Start`.
- [x] 1.3 Add failing docs guard for the public P2P CLI startup cancellation contract.
- [x] 1.4 Add failing coverage proving ephemeral P2P CLI cleanup waits for startup workers.

## 2. Implementation

- [x] 2.1 Change `internal/p2p.Node.Start` to derive its lifecycle context from a caller context.
- [x] 2.2 Change P2P CLI dependency initialization and all affected loader seams to accept command context.
- [x] 2.3 Preserve existing cleanup and command output behavior.

## 3. Downstream Artifacts

- [x] 3.1 Update public P2P CLI docs with the verified startup cancellation behavior.
- [x] 3.2 Sync OpenSpec main specs for P2P networking, CLI management, docs, and test coverage.

## 4. Verification

- [x] 4.1 Run focused P2P CLI and docs guard tests.
- [x] 4.2 Run `go build ./...`, `go test ./...`, `git diff --check`, and strict OpenSpec validation.
- [x] 4.3 Run subagent-driven review and address required findings.
- [x] 4.4 Archive the change and commit the scoped unit.
