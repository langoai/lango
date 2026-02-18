## Context

Chat Response Speed Optimization has been implemented across several packages (gateway streaming events, embedding query cache, observational memory context limits). The README.md needs to reflect these new capabilities and configuration options.

## Goals / Non-Goals

**Goals:**
- Document all new configuration fields in the reference table
- Add WebSocket event documentation for frontend developers
- Document the embedding cache behavior
- Document context limit behavior for observational memory

**Non-Goals:**
- No code changes
- No new features or configuration options
- No changes to existing documentation structure

## Decisions

1. **WebSocket Events section placement**: Add under the existing WebSocket CORS section as a sibling `####` heading, keeping all WebSocket-related docs grouped together.
2. **Embedding Cache**: Add as a `###` subsection under Embedding & RAG, since it's a self-contained capability within that feature area.
3. **Context Limits**: Add as a bullet point in the existing Observational Memory component list, maintaining the established pattern.

## Risks / Trade-offs

- [Minimal risk] Documentation-only change with no runtime impact.
