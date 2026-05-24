## Why

The inline approval strip is documented as a compact single-line prompt, but its summary text is still rendered without stripping terminal control sequences or collapsing embedded newlines first. A malformed tool summary can still break that compact surface.

## What Changes

- Sanitize inline approval-strip summary text with ANSI/OSC stripping plus single-line normalization.
- Add regression coverage for multiline and escaped summaries.
- Record the compact-summary contract in OpenSpec and downstream cockpit docs.

## Impact

- Keeps Tier 1 approval prompts aligned with the same single-line rendering baseline as the rest of chat.
- Prevents malformed tool summaries from destabilizing the inline approval surface.
