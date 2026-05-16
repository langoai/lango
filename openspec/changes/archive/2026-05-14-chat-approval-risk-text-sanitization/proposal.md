## Why

Approval surfaces already sanitize tool names, summaries, and channel-origin text, but the fullscreen approval dialog still renders `Risk.Label` and `RuleExplanation` raw. Malformed approval metadata can still inject control sequences or break the dialog's text layout.

## What Changes

- Sanitize fullscreen approval-dialog risk-label and rule-explanation text.
- Add regression coverage for escaped and multiline risk metadata.
- Record the risk-text sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the approval-surface plain-text baseline to the last visible metadata fields in the dialog.
- Prevents malformed approval risk metadata from destabilizing Tier 2 UI.
