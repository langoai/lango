## 1. Output Routing

- [x] 1.1 Route `lango mcp list` table output through the Cobra command writer.
- [x] 1.2 Route the empty-state guidance through the Cobra command writer.
- [x] 1.3 Add command-level capture tests for configured and empty states.

## 2. Spec Sync

- [x] 2.1 Record the MCP list output-writer contract in `mcp-integration`.
- [x] 2.2 Update downstream `docs/cli/mcp.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/mcp -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-mcp-list-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
