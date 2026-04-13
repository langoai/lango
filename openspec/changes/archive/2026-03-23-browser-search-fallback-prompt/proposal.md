## Why

The navigator prompt currently over-emphasizes `browser_search` as the preferred path for live web queries, but it does not tell the model what to do when that tool is unavailable in the current runtime. That creates avoidable dead-ends even when lower-level browser navigation and extraction tools are still available.

## What Changes

- Add explicit browser fallback guidance to `TOOL_USAGE.md`.
- Add explicit navigator fallback protocol to the navigator sub-agent prompts.
- Instruct the navigator to fall back from `browser_search` to `browser_navigate` plus `browser_extract(search_results)`, and then to low-level `browser_action`/`eval` when required.
- Clarify that missing higher-level tools should not cause the navigator to stop if equivalent lower-level browser tools remain available.

## Capabilities

### New Capabilities

### Modified Capabilities

- `agent-prompting`: Prompt guidance for browser workflows now includes an explicit fallback chain when higher-level browser tools are unavailable.
- `multi-agent-orchestration`: Navigator sub-agent instructions now define a browser fallback protocol instead of failing early on missing high-level tools.

## Impact

- `prompts/TOOL_USAGE.md`
- `prompts/agents/navigator/IDENTITY.md`
- `internal/agentregistry/defaults/navigator/AGENT.md`
- `openspec/specs/agent-prompting/spec.md`
- `openspec/specs/multi-agent-orchestration/spec.md`
