## 1. Output Routing

- [x] 1.1 Route `lango agent list` table output through the Cobra command writer.
- [x] 1.2 Route `lango agent list --json` output through the Cobra command writer.
- [x] 1.3 Route empty/filter-miss and remote-section spacing through the Cobra command writer.
- [x] 1.4 Add command-level capture tests for representative list flows.

## 2. Spec Sync

- [x] 2.1 Record the list output-writer contract in `cli-agent-inspection`.
- [x] 2.2 Update downstream `docs/cli/agent-memory.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/agent -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-agent-list-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
