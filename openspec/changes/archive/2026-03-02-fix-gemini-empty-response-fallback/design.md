## Context

Gemini 3 models use a "thinking" mechanism where intermediate reasoning steps are returned as text parts with `Thought=true`. The current provider filters these with `!part.Thought`, which works for mixed responses (thought + visible text) but fails completely for thought-only responses — all text is discarded, resulting in empty strings propagated to channels. The empty response guard added previously prevents API errors but leaves users without any feedback.

Additionally, `agent.go` contains `!part.Thought` filters on session event parts that are dead code — `model.go` never sets `Thought=true` on the `genai.Part` objects it creates from `StreamEventPlainText` events.

## Goals / Non-Goals

**Goals:**
- Users ALWAYS receive a response, even when the model produces only thought text
- Thought text becomes observable (logged with length) instead of silently dropped
- Dead code removed from agent.go for clarity
- Both channel (Telegram/Discord/Slack) and gateway (WebSocket) paths protected

**Non-Goals:**
- Surfacing thought content to users (privacy/UX concern — only length is logged)
- Changing how thought tool calls are handled (those already work correctly)
- Modifying Gemini API request parameters to prevent thought-only responses

## Decisions

**1. Single fallback point per entry path**
- Channel path: `runAgent` in `channels.go` — all three channels (Telegram/Discord/Slack) converge here
- Gateway path: `handleChatMessage` in `server.go` — WebSocket streaming path
- Alternative: per-channel fallback → rejected (DRY violation, easy to miss one)

**2. Observable thought events instead of silent drop**
- New `StreamEventThought` type emitted by Gemini provider with `ThoughtLen` metadata
- Alternative: log inside gemini.go directly → rejected (breaks separation of concerns, provider shouldn't know about logging framework)

**3. Dead code removal over annotation**
- `!part.Thought` in agent.go can never trigger because model.go creates text Parts without setting Thought
- Alternative: keep with `// NOTE: currently unreachable` → rejected (misleading, adds confusion)

## Risks / Trade-offs

- [Fallback message is generic] → Acceptable for now; specific guidance ("rephrase") is actionable
- [ThoughtLen leaks thought existence] → Minimal risk; length metadata is useful for diagnostics without exposing content
- [Dead code removal could break if model.go changes] → Thought filtering is now handled at provider level, not agent level; this is the correct architectural boundary
