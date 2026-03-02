## Why

Gemini 3 models produce thought-only responses where all text parts have `Thought=true`. The provider's `!part.Thought` filter silently discards all text, resulting in `response_len: 0` reaching channels. While the empty response guard prevents Telegram API errors, users receive **no response at all** — worse UX than v0.3.0. Additionally, `agent.go` contains dead code filtering on `!part.Thought` that can never trigger because `model.go` never sets `Thought=true` on text parts.

## What Changes

- Add fallback message in `channels.go:runAgent` when agent returns empty string — ensures all channel users always get a response
- Add identical fallback in `gateway/server.go:handleChatMessage` for WebSocket streaming path
- Introduce `StreamEventThought` event type in provider interface — thought text is now observable instead of silently dropped
- Modify `gemini.go` to emit `StreamEventThought` events with length metadata instead of discarding thought text
- Add explicit `StreamEventThought` handling in `model.go` (no-op, prevents unhandled case)
- Remove dead `!part.Thought` filter from `agent.go` (4 locations) — these conditions could never be true
- Add warn-level logging when agent returns empty response for diagnostics

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `provider-interface`: Adding `StreamEventThought` event type and `ThoughtLen` field to `StreamEvent`
- `gemini-content-sanitization`: Thought text now emitted as observable event instead of silent drop
- `session-store`: No schema change, but empty response fallback affects what gets stored

## Impact

- `internal/app/channels.go` — new constant + guard logic
- `internal/gateway/server.go` — new constant + guard logic
- `internal/provider/provider.go` — new event type + struct field
- `internal/provider/gemini/gemini.go` — thought event emission
- `internal/adk/model.go` — thought event handling
- `internal/adk/agent.go` — dead code removal + warn logging
