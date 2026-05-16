## Why

The chat header is now nil-safe and width-safe, but it still renders provider, model, and session key text without stripping control sequences or collapsing embedded whitespace. Corrupted or malformed values can still break the single-line shell-bar contract.

## What Changes

- Sanitize provider, model, and session key display text with ANSI/OSC stripping plus single-line normalization.
- Add regression coverage for malformed header field values.
- Record the shell-header field-sanitization contract in OpenSpec.

## Impact

- Hardens the shell header against malformed config/session metadata.
- Keeps the fixed shell bar aligned with the rest of the single-line rendering baseline.
