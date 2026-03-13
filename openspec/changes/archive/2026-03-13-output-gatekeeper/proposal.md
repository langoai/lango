## Why

The agent's observation and response channels share the same output path, causing the LLM to echo raw tool results, internal reasoning tags, and bulk JSON directly to users. This degrades user experience with noisy, confusing responses. A code-level isolation layer is needed between what the model sees and what the user receives.

## What Changes

- Add a **tool output truncation middleware** that caps tool result text before it enters model context (default 8000 chars)
- Add a **response sanitizer** that strips thought/thinking tags, internal markers, and oversized JSON from model output
- Add **system prompt output principles** instructing the model to summarize rather than echo raw data
- New `gatekeeper` config section with per-rule toggle switches (`*bool` pointer pattern, enabled by default)
- New `tools.maxOutputChars` config field for truncation limit

## Capabilities

### New Capabilities
- `output-gatekeeper`: 3-layer response filtering — preventive (system prompt), upstream (tool truncation), downstream (response sanitization)

### Modified Capabilities
- `tool-execution`: Tool results now pass through truncation middleware before entering model context
- `prompt-system`: System prompt gains a new "Output Principles" section at priority 350

## Impact

- **Core**: `internal/toolchain/` (new middleware), `internal/gatekeeper/` (new package), `internal/prompt/` (new section)
- **Config**: `internal/config/types.go` — new `GatekeeperConfig`, `MaxOutputChars` field
- **App wiring**: `internal/app/app.go`, `channels.go`, `types.go` — sanitizer init + response filtering
- **Gateway**: `internal/gateway/server.go` — chunk and final response sanitization
- **Prompts**: `prompts/OUTPUT_PRINCIPLES.md` — embedded prompt file
