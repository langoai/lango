## Why

Channel badge text is now sanitized, but the badge color still keys off the raw `channel` string. If the incoming channel name contains control sequences around an otherwise valid channel label, the badge text and badge color can disagree.

## What Changes

- Sanitize the channel key used for badge color selection.
- Add regression coverage for escaped known channel names.
- Record the sanitized color-key contract in OpenSpec and downstream cockpit docs.

## Impact

- Keeps channel badge text and color mapping aligned under malformed input.
- Removes the last channel-row display mismatch between sanitized text and raw metadata.
