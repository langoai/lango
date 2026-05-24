## Why

Approval surfaces already sanitize displayed channel-origin values and badge text, but they still decide whether a session key is Telegram/Discord/Slack from the raw prefix segment. An escaped known prefix can therefore miss the known-channel path entirely even though the visible text is otherwise sanitized.

## What Changes

- Sanitize the channel-prefix segment before origin/badge channel matching.
- Add regression coverage for escaped known session-key prefixes.
- Record the prefix-sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Keeps approval origin/badge extraction aligned with the rest of the sanitized channel metadata path.
- Prevents malformed session-key prefixes from silently dropping known-channel affordances.
