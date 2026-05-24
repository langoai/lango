## 1. Tests

- [x] 1.1 Add failing tests for command-context cancellation reaching host connect.
- [x] 1.2 Add failing tests for configured handshake timeout and 30 second fallback timeout selection.
- [x] 1.3 Add failing test for cleanup after connect timeout/failure.

## 2. Implementation

- [x] 2.1 Thread `cmd.Context()` into the P2P connect path.
- [x] 2.2 Apply bounded timeout selection from `p2p.handshakeTimeout`, with 30 second fallback.
- [x] 2.3 Preserve cleanup on parse and connect failures.
- [x] 2.4 Return actionable peer-scoped timeout/cancellation errors.

## 3. Downstream Artifacts

- [x] 3.1 Update public P2P CLI docs with the verified bounded connect behavior.
- [x] 3.2 Sync OpenSpec main specs for P2P connect, docs, and test coverage.

## 4. Verification

- [x] 4.1 Run focused P2P connect tests.
- [x] 4.2 Run `go build ./...`, `go test ./...`, `git diff --check`, and strict OpenSpec validation.
- [x] 4.3 Run subagent-driven review and address required findings.
- [x] 4.4 Archive the change and commit the scoped unit.
