## Why

Channel transcript rows already sanitize sender and message text, but the badge label itself still renders the `channel` string raw. A malformed external channel name can still inject control sequences into the badge.

## What Changes

- Sanitize displayed channel badge text before rendering it.
- Add regression coverage for escaped and multiline channel names.
- Record the channel-badge sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Completes the plain-text rendering baseline for channel transcript rows.
- Prevents malformed channel labels from destabilizing the badge surface.
