## 1. Output Routing

- [x] 1.1 Route `lango graph query` text output through the Cobra command writer.
- [x] 1.2 Route `lango graph query --json` output through the Cobra command writer.
- [x] 1.3 Add command-level capture tests for representative query flows.

## 2. Spec Sync

- [x] 2.1 Record the graph query output-writer contract in `cli-graph-management`.
- [x] 2.2 Update downstream `docs/cli/agent-memory.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/graph -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-graph-query-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
