## Context

Telegram Markdown v1 uses `*` for bold, `_` for italic, and `` ` `` for code. The `FormatMarkdown` function in `internal/channels/telegram/format.go` converts standard Markdown `**bold**` to `*bold*`. When the LLM generates `*`-style bullet lists (e.g., `* **작업 이름:** ...`), the conversion produces lines with an odd number of `*` characters, causing Telegram's parser to reject the message with "Can't find end of the entity" errors.

The existing `Send()` method already has a plain text fallback that retries without formatting when the API rejects a message.

## Goals / Non-Goals

**Goals:**
- Prevent malformed Markdown at the source by instructing the LLM to follow channel-safe formatting conventions
- Eliminate the most common parse error patterns: `*` bullets, bare underscores, unpaired markers

**Non-Goals:**
- Rewriting the `FormatMarkdown` converter to handle all edge cases programmatically
- Adding MarkdownV2 support (different scope)
- Modifying the plain text fallback mechanism

## Decisions

**Decision 1: Prompt-level fix over code-level fix**
- Rationale: The root cause is LLM-generated Markdown that is valid standard Markdown but incompatible with Telegram v1. Instructing the LLM to avoid problematic patterns is simpler, lower-risk, and covers all channels at once. The existing fallback remains as a safety net.
- Alternative considered: Fixing `FormatMarkdown` to escape lone `*`, `_` characters. Rejected because the number of edge cases is unbounded and the formatter would grow increasingly complex.

**Decision 2: Rules added to CONVERSATION_RULES.md**
- Rationale: This file already contains formatting guidelines (line 8: "Use appropriate formatting") and channel limits (line 5). Adding Markdown conventions here keeps all response formatting rules in one place.
- Alternative considered: Adding to AGENTS.md identity section. Rejected because formatting rules belong with other conversation/formatting guidance.

## Risks / Trade-offs

- [LLM non-compliance] LLMs may occasionally ignore prompt instructions → Mitigated by the existing plain text fallback in `Send()`.
- [Prompt length] Additional rules increase system prompt size → Impact is minimal (5 lines added).
