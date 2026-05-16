## Why

Tool names are still rendered raw across approval and tool-lifecycle surfaces. If a remote tool catalog or external approval source provides a malformed name containing ANSI/OSC escape sequences or embedded newlines, the chat UI can still display unsafe or layout-breaking text even though summaries and origins are already sanitized.

## What Changes

- Sanitize displayed tool names across approval and tool-lifecycle surfaces.
- Add regression coverage for escaped and multiline tool names.
- Record the tool-name sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the plain-text rendering baseline from summaries/origins to tool names themselves.
- Prevents malformed remote tool names from destabilizing chat approval or lifecycle UI.
