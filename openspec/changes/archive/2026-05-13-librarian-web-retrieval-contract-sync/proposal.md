## Why

The runtime already routes `web_search` and `web_fetch` to the librarian specialist, but the current librarian prompts still say "Never browse the web" and the public multi-agent docs omit `web_*` from the librarian role. The web tool handlers also lacked direct entrypoint regressions for missing `query` and `url`.

## What Changes

- Add direct tool-entrypoint regressions for missing `query` on `web_search` and missing `url` on `web_fetch`.
- Update the librarian runtime prompt and embedded prompt file to describe lightweight web retrieval explicitly.
- Sync TOOL_USAGE, README, and multi-agent docs with the librarian `web_*` routing and required-input contract.
- Sync web tool, agent-prompting, downstream-docs, and production-readiness specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `tool-websearch` and `tool-webfetch`: required-input guards are now directly covered at the tool entrypoint.
- `multi-agent-orchestration`: librarian prompt wording now explicitly covers lightweight web retrieval.
- `downstream-docs-sync`: librarian role docs now mention `web_search` and `web_fetch`.
- `production-readiness`: wrapper guard coverage now includes lightweight web retrieval tools.

## Impact

- Affected code: `internal/orchestration/tools.go`
- Affected tests: `internal/tools/websearch/tools_test.go`, `internal/tools/webfetch/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`, `prompts/agents/librarian/IDENTITY.md`
- Affected docs: `README.md`, `docs/features/multi-agent.md`
- Affected specs: `openspec/specs/tool-websearch/spec.md`, `openspec/specs/tool-webfetch/spec.md`, `openspec/specs/agent-prompting/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
