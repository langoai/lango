## Why

The navigator no longer loops on identical approval prompts, but it still fails to converge after approval because it keeps reformulating searches and reissuing browser queries instead of working from the current results page. The browser tools also do not expose strong enough "stop searching" signals for the model to know when it already has sufficient results.

## What Changes

- Strengthen navigator/browser prompt guidance to enforce a bounded search workflow: search once, work from the current page, at most one reformulation, and stop when the requested result count is satisfied.
- Enrich `browser_navigate`, `browser_search`, and `browser_extract(search_results|article)` outputs with page-type and result-count signals.
- Add request-local browser search churn diagnostics so repeated search reformulation is visible in logs.
- Preserve the current approval policy and exact-match replay guard behavior.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tool-browser`: Browser search and extraction outputs now carry explicit convergence signals and request-local search churn diagnostics.
- `agent-prompting`: Browser guidance now enforces search-once/extract-first bounded search behavior.
- `multi-agent-orchestration`: Navigator instructions now prefer current-page extraction over repeated search and enforce a single reformulation budget.

## Impact

- `internal/tools/browser/`
- `internal/app/channels.go`
- `internal/gateway/server.go`
- `prompts/TOOL_USAGE.md`
- `prompts/agents/navigator/IDENTITY.md`
- `internal/agentregistry/defaults/navigator/AGENT.md`
- `README.md`
- `docs/features/multi-agent.md`
