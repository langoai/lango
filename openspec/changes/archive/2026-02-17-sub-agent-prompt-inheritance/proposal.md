## Why

In multi-agent mode, sub-agents (operator, navigator, vault, librarian, planner, chronicler) only receive hard-coded instructions from `agentSpecs`. They do not inherit shared prompt sections (SAFETY.md, CONVERSATION_RULES.md) that the user has configured via `promptsDir`. This means safety guidelines and conversation rules are inconsistently applied — the orchestrator follows them but sub-agents do not.

## What Changes

- Add `Builder.Clone()` method to prompt builder for independent branching per sub-agent
- Add `SectionAgentIdentity` section ID (priority 150) to distinguish sub-agent role from global identity
- Add `LoadAgentFromDir()` to load per-agent prompt overrides from `agents/<name>/` subdirectories
- Add `SubAgentPromptFunc` callback type to orchestration Config for prompt assembly injection
- Wire `buildSubAgentPromptFunc()` in app wiring to assemble sub-agent prompts with shared Safety + ConversationRules + per-agent overrides
- Update Settings TUI hint text for Prompts Directory field
- Document per-agent prompt customization in README

## Capabilities

### New Capabilities
- `sub-agent-prompt-inheritance`: Sub-agents inherit shared prompt sections and support per-agent customization via `agents/<name>/` directories

### Modified Capabilities
- `structured-prompt-builder`: Add `Clone()` method and `SectionAgentIdentity` constant
- `multi-agent-orchestration`: Add `SubAgentPromptFunc` to Config and use it in `BuildAgentTree`

## Impact

- `internal/prompt/builder.go` — new `Clone()` method
- `internal/prompt/section.go` — new `SectionAgentIdentity` constant
- `internal/prompt/loader.go` — new `agentSectionFiles` map and `LoadAgentFromDir()` function
- `internal/orchestration/orchestrator.go` — new `SubAgentPromptFunc` type, added to Config, used in BuildAgentTree
- `internal/app/wiring.go` — new `buildSubAgentPromptFunc()`, wired into orchestration Config
- `internal/cli/settings/forms_impl.go` — updated placeholder text
- `README.md` — new "Per-Agent Prompt Customization" documentation section
