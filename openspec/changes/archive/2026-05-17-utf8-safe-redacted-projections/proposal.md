## Why

Redacted plaintext projections are stored for search and display while original sensitive payloads are encrypted. The projection truncation path currently slices by byte length, which can split a multi-byte UTF-8 rune and persist invalid text for non-English user content.

## What Changes

- Add regression tests for redaction, whitespace normalization, and UTF-8-safe truncation.
- Update `RedactedProjection` to keep byte-limit behavior while trimming to the last valid UTF-8 boundary.
- Document that projection limits never split UTF-8 characters.

## Impact

- Improves production robustness for multilingual user content.
- No storage schema or encryption format changes.
- Existing byte limits remain upper bounds.
