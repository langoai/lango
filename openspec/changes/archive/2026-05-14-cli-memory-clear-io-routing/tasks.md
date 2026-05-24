## 1. Command Streams

- [x] 1.1 Route `lango memory clear` prompt and result output through the Cobra command writer.
- [x] 1.2 Read `lango memory clear` confirmation input through the Cobra command input reader.
- [x] 1.3 Add command-level capture tests for abort, confirm, and force flows.

## 2. Spec Sync

- [x] 2.1 Record the memory clear I/O routing contract in `cli-memory-management`.
- [x] 2.2 Update downstream `docs/cli/agent-memory.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/memory -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-memory-clear-io-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
