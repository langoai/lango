## Tasks

- [x] 1.1 Add isolated child overlay tracking to `SessionServiceAdapter`
- [x] 1.2 Write isolated events to child history and parent in-memory overlay, but not to parent persistent store
- [x] 1.3 Roll back overlay before success merge/discard finalization
- [x] 1.4 Keep success path root-authored summary-only and add discard failure notes
- [x] 2.1 Add regression tests for isolated overlay visibility and function-response replay
- [x] 2.2 Add regression tests for summary merge and discard-note cleanup behavior
- [x] 2.3 Keep non-isolated behavior unchanged
- [x] 3.1 Update architecture/user docs to the new semantics
- [x] 3.2 Sync main specs for `sub-session-isolation` and `multi-agent-orchestration`
- [x] 4.1 Run `go build ./...`
- [x] 4.2 Run `go test ./...`
- [x] 4.3 Archive the change
