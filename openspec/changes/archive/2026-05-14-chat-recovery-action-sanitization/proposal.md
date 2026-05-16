## Why

The recovery transcript row now sanitizes `causeClass`, but `action` still feeds the display-name mapping raw. A malformed action string containing control sequences or embedded newlines can silently fall off the known-action path and render the generic `Recovery` label.

## What Changes

- Sanitize the recovery action string before mapping it to its display label.
- Add regression coverage for escaped known recovery actions.
- Record the sanitized recovery-action contract in OpenSpec and downstream cockpit docs.

## Impact

- Keeps recovery action label mapping aligned with the rest of the sanitized recovery metadata path.
- Prevents malformed action values from silently degrading the visible recovery label.
