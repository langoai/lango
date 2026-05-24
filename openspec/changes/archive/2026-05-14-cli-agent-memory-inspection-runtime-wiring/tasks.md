## 1. Runtime Wiring

- [x] 1.1 Wire `lango memory agents` to the persistent agent memory store and render real summaries.
- [x] 1.2 Wire `lango memory agent <name>` to the persistent agent memory store and render real entries.
- [x] 1.3 Add `--limit` support for `lango memory agent <name>`.
- [x] 1.4 Add command-level tests for table and JSON inspection flows.

## 2. Spec Sync

- [x] 2.1 Sync `cli-memory-management` with the runtime-backed agent memory inspection behavior.
- [x] 2.2 Update downstream `docs/cli/agent-memory.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/memory -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-agent-memory-inspection-runtime-wiring --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
