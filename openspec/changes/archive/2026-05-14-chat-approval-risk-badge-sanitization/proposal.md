## Why

The fullscreen approval dialog already sanitizes tool names, summaries, origin text, risk labels, and `Why:` explanations, but the risk badge text still comes from raw `Risk.Level` with only `strings.ToUpper(...)`. A malformed level string can still inject control sequences into the badge.

## What Changes

- Sanitize fullscreen approval-dialog risk-badge text before uppercasing/rendering it.
- Add regression coverage for escaped and multiline risk levels.
- Record the risk-badge sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Completes the plain-text rendering baseline for all visible risk metadata in the Tier 2 approval dialog header.
- Prevents malformed risk-level values from destabilizing the badge surface.
