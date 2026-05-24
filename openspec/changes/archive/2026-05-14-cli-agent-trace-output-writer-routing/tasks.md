## 1. Output Routing

- [x] 1.1 Route `lango agent trace list` table and JSON output through the Cobra command writer.
- [x] 1.2 Route `lango agent trace show` table and JSON output through the Cobra command writer.
- [x] 1.3 Add command-level capture tests backed by a seeded trace store.

## 2. Spec Sync

- [x] 2.1 Record the trace list/show output-writer contract in `cli-agent-inspection`.
- [x] 2.2 Update downstream `docs/cli/core.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/agent -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-agent-trace-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
