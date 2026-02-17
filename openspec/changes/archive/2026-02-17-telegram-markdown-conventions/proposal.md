## Why

Telegram Markdown v1 parser rejects messages containing malformed markers (unpaired `*`, `_`, `` ` ``), causing `"Can't find end of the entity"` errors. The root cause is LLM-generated responses using `*` for bullet lists, which conflict with `**bold**` → `*bold*` conversion in `FormatMarkdown`, producing odd counts of `*` that Telegram cannot pair. Adding strict Markdown formatting rules to the agent's conversation prompt prevents malformed output at the source.

## What Changes

- Add a "Follow strict Markdown conventions" rule to `prompts/CONVERSATION_RULES.md` with sub-rules:
  - Always use `-` for unordered lists, never `*`
  - Wrap code identifiers containing underscores in backticks
  - Ensure all formatting markers (`*`, `_`, `` ` ``) are paired
  - Always close fenced code blocks

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `channel-telegram`: Add requirement that the agent prompt includes Markdown formatting conventions to prevent v1 parse errors.
- `embedded-prompt-files`: Add requirement that CONVERSATION_RULES.md includes channel-safe Markdown formatting guidelines.

## Impact

- `prompts/CONVERSATION_RULES.md` — updated with new formatting rules
- All channels benefit (Telegram, Discord, Slack) since the rules enforce clean Markdown universally
- No code changes required — prompt-only change
- Existing plain text fallback in `Send()` remains as safety net
