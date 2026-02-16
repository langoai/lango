## Why

The Telegram bot accumulates and repeats previous answers when responding to new questions. The root cause is that the system prompt has no conversation behavior guidelines and the entire conversation history (up to 50 turns) is sent to the LLM each time. The current `_defaultSystemPrompt` is a single-line string with no extensibility. A structured, section-based prompt builder is needed to inject conversation rules by default and allow per-section customization via `.md` files.

## What Changes

- New `internal/prompt` package with `PromptSection` interface, `StaticSection` type, `Builder`, default sections, and directory-based `.md` loader
- Four built-in prompt sections: Identity (100), Safety (200), Conversation Rules (300), Tool Usage (400)
- Directory-based prompt customization: users place `.md` files in a configurable `promptsDir` to override individual sections
- `AgentConfig` gains a `PromptsDir` field for the prompts directory path
- `ContextAwareModelAdapter` constructor accepts `*prompt.Builder` instead of a raw `string`
- `wiring.go` replaces `loadSystemPrompt()` with `buildPromptBuilder()` supporting three modes: directory-based, legacy single-file, and built-in defaults

## Capabilities

### New Capabilities
- `structured-prompt-builder`: Section-based system prompt construction with priority ordering, default sections, and file-based override/customization

### Modified Capabilities
- `agent-prompting`: System prompt construction now uses a structured builder instead of a single string; conversation rules are included by default
- `agent-provider-config`: `AgentConfig` adds `PromptsDir` field for prompt directory configuration

## Impact

- `internal/prompt/` — new package (section.go, sections.go, builder.go, defaults.go, loader.go + tests)
- `internal/config/types.go` — `AgentConfig.PromptsDir` field added
- `internal/adk/context_model.go` — constructor signature changed from `string` to `*prompt.Builder`
- `internal/app/wiring.go` — `_defaultSystemPrompt` and `loadSystemPrompt()` replaced with `buildPromptBuilder()`
- Backward compatible: existing `SystemPromptPath` config still works; no config = defaults with conversation rules
