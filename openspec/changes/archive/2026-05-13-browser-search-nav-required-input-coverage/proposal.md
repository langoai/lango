## Why

The browser tool surface already declares `query` as required for `browser_search` and `url` as required for `browser_navigate`, but there was no direct regression proving that those missing-input failures happen before session creation or navigation. The browser prompt and operator-facing docs also did not state that top-level contract explicitly.

## What Changes

- Add browser tool regressions for missing `query` and `url` at the search/navigation entrypoint.
- Sync browser prompts and public multi-agent docs with the top-level browser input contract.
- Sync browser, agent-prompting, downstream-docs, and production-readiness specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `tool-browser`: browser search/navigation required-input guards are now explicitly covered.
- `agent-prompting`: browser guidance now states the top-level `query`/`url` contract.
- `downstream-docs-sync`: browser docs now mention the search/navigation required inputs.
- `production-readiness`: coverage now includes browser search/navigation wrapper guards.

## Impact

- Affected tests: `internal/tools/browser/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`, `prompts/agents/navigator/IDENTITY.md`
- Affected docs: `README.md`, `docs/cli/agent-memory.md`
- Affected specs: `openspec/specs/tool-browser/spec.md`, `openspec/specs/agent-prompting/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
