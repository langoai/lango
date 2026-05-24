## Why

The graph and persistent agent-memory tools already reject missing required inputs such as `start_node`, `query`, `key`, and `content`. But there was no direct tool-entrypoint regression proving those failures happen before traversal, lookup, or mutation, and the shared prompt plus public multi-agent docs did not state that contract explicitly.

## What Changes

- Add direct tool-entrypoint regressions for missing required inputs across graph and agent-memory tools.
- Update TOOL_USAGE, README, and multi-agent docs to describe the graph and persistent-memory required-input contract.
- Sync agent-prompting, downstream-docs, and production-readiness specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `graph` and `agentmemory`: required-input guards are now directly covered at the tool entrypoint.
- `downstream-docs-sync`: graph and agent-memory docs now mention the required wrapper inputs.
- `production-readiness`: wrapper guard coverage now includes graph and agent-memory tools.

## Impact

- Affected tests: `internal/graph/tools_test.go`, `internal/agentmemory/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`
- Affected docs: `README.md`, `docs/features/multi-agent.md`
- Affected specs: `openspec/specs/agent-prompting/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
