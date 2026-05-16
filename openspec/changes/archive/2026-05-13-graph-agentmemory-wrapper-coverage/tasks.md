## 1. Guard Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for graph required inputs.
- [x] 1.2 Add tool-entrypoint regression coverage for agent-memory required inputs.

## 2. Prompt And Docs Sync

- [x] 2.1 Update TOOL_USAGE for graph and agent-memory required-input guidance.
- [x] 2.2 Update README and multi-agent docs for the same contract.
- [x] 2.3 Update agent-prompting, downstream-docs, and production-readiness specs for the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/graph ./internal/agentmemory -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
