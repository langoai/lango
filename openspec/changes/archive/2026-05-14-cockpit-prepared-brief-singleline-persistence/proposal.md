## Why

The durable proposal-description sanitization path strips control sequences from prepared-brief segments, but `compactPreparedBrief()` still joins those segments with embedded newlines. That means accepted proposals can still persist multiline mission descriptions even though the contract now requires plain single-line durable text.

## What Changes

- Collapse prepared-brief segments into one single-line persisted description.
- Add regression coverage to ensure accepted proposal descriptions no longer retain embedded newlines.
- Record the single-line prepared-brief persistence contract in OpenSpec and downstream docs.

## Impact

- Closes the remaining multiline persistence gap in accepted proposal descriptions.
- Keeps durable mission descriptions aligned with the same replay-safe baseline as other cockpit text.
