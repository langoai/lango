## Why

Browser-heavy requests are currently forced through low-level `browser_*` primitives, which burns turn budget quickly and makes simple live-web tasks unreliable. When a run fails, raw partial text from the agent stack can also leak back to users, which blurs the boundary between internal orchestration and user-facing output.

## What Changes

- Raise the baseline agent turn budget so non-trivial tool workflows do not fail prematurely.
- Raise the implicit multi-agent turn budget again to keep extra headroom over the new baseline.
- Upgrade browser tooling from page-only primitives to richer research primitives: structured navigation snapshots, browser-native search, page observation, and structured extraction.
- Stop surfacing raw partial drafts to users on timeout/turn-limit failures; keep them internal for diagnostics only.
- Stop persisting raw partial drafts into session history timeout annotations.
- Tighten output principles so prompts explicitly forbid dumping system/user/tool framing into user-visible replies.
- Update CLI inspection output, prompts, README, and docs to reflect the new defaults and browser capabilities.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tool-browser`: Add higher-level browser search/observe/extract flows and richer navigation snapshots.
- `agent-turn-limit`: Increase the standard default turn limit from 25 to 50.
- `multi-agent-orchestration`: Increase the implicit multi-agent turn limit from 50 to 75.
- `agent-error-handling`: Preserve partial drafts internally but stop exposing raw partial text to users.
- `adaptive-idle-timeout`: Timeout annotations no longer persist raw partial drafts into session history.
- `output-gatekeeper`: Output principles explicitly forbid role-labeled prompt/tool dumps in user-visible replies.
- `config-types`: Zero-value `agent.maxTurns` defaults change to the new single-agent and multi-agent effective values.
- `cli-agent-inspection`: `lango agent status` reports the new effective turn defaults.

## Impact

- `internal/adk/`
- `internal/tools/browser/`
- `internal/app/`
- `internal/gateway/`
- `internal/toolchain/`
- `internal/config/`
- `internal/cli/agent/`
- `prompts/`
- `README.md`
- `docs/`
