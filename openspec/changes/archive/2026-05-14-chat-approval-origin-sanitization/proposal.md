## Why

Approval surfaces already sanitize summaries and params, but channel-origin display text still renders the session-key origin segment raw. A malformed or corrupted session key can still inject control sequences or embedded newlines into the origin line.

## What Changes

- Sanitize displayed channel-origin text with ANSI/OSC stripping and single-line normalization.
- Add regression coverage for malformed channel-origin session keys.
- Record the origin-display contract in OpenSpec and downstream cockpit docs.

## Impact

- Keeps channel-origin affordances aligned with the rest of the hardened approval surfaces.
- Prevents malformed session keys from destabilizing banner/dialog origin lines.
