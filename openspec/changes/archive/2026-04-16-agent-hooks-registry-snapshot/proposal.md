## Why

`lango agent hooks` currently displays only config flags (enabled/disabled booleans from `HooksConfig`). It does not reveal which hooks are actually registered in the runtime `HookRegistry`, their priority order, or their wired state (e.g. which tools `KnowledgeSaveHook` is configured to save). Operators and developers diagnosing hook behavior must read source code or app logs instead of using a CLI command. This change completes the existing hook visibility surface so it reflects actual runtime state.

## What Changes

- Extend `lango agent hooks` to additionally display the **registry snapshot**: registered pre/post hooks with name, priority, phase (pre/post), and wired details (e.g. `KnowledgeSaveHook.SaveableTools` list).
- Export a public `BuildHookRegistry(cfg) → *HookRegistry` helper from `internal/app/` so the CLI can produce a registry snapshot without a full app bootstrap.
- Add a default allowlist constant for `KnowledgeSaveHook.SaveableTools` (read-type tools) and wire it in `app.go` builder.
- Preserve full backward compatibility: existing config-only JSON/text output fields remain unchanged; new fields are additive.

## Capabilities

### New Capabilities
_(none — this extends an existing capability)_

### Modified Capabilities
- `cli-agent-tools-hooks`: The "Agent hooks command" requirement currently mandates "cfgLoader (config only)". This change extends it to also produce a registry snapshot via a public builder helper, while keeping cfgLoader as the config source.

## Impact

- `internal/app/app.go` — extract `buildHookRegistry` to a public export `BuildHookRegistry`
- `internal/toolchain/hook_knowledge.go` — add `DefaultSaveableTools` constant
- `internal/cli/agent/hooks.go` — consume registry snapshot, extend output structs
- `internal/cli/settings/forms_hooks.go` — no config field added, minimal or no change (SaveableTools is a code constant, not a config field)
- `openspec/specs/cli-agent-tools-hooks/spec.md` — delta spec extending the hooks command requirement
