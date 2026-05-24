## Why

The browser tool spec already says each `browser_action` mode depends on different required inputs, but the codebase did not have direct regressions for those action-specific validation boundaries. The prompt and public multi-agent docs also did not state that contract explicitly, which left a gap between runtime behavior, agent guidance, and operator-facing documentation.

## What Changes

- Add browser tool regressions for missing action-specific `selector`/`text` inputs.
- Add regression coverage for the P2P `eval` block happening before session creation.
- Sync browser prompt and public multi-agent docs with the action-specific `browser_action` contract.
- Sync browser and production-readiness specs with the conditional-input guard coverage.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `tool-browser`: `browser_action` conditional required-input guards are now explicitly covered and documented.
- `agent-prompting`: browser guidance now states the action-specific `browser_action` input contract.
- `downstream-docs-sync`: browser tool docs now mention the `browser_action` input contract in README and CLI docs.
- `production-readiness`: regression coverage now includes browser action conditional input guards.

## Impact

- Affected tests: `internal/tools/browser/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`, `prompts/agents/navigator/IDENTITY.md`
- Affected docs: `README.md`, `docs/cli/agent-memory.md`
- Affected specs: `openspec/specs/tool-browser/spec.md`, `openspec/specs/agent-prompting/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
