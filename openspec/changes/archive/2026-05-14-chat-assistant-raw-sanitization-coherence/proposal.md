## Why

Assistant markdown rendering now strips terminal control sequences before display, but the stored assistant `rawContent` and duplicate-failure suppression path still operate on the raw unsanitized model output. That leaves model storage and fallback/dedup logic out of sync with the rendered transcript.

## What Changes

- Strip terminal control sequences from stored assistant raw content while preserving markdown/newline structure.
- Use the same stripped text when comparing failure text against the last assistant content for duplicate suppression.
- Add regression coverage for sanitized assistant raw storage and duplicate suppression.
- Record the coherence contract in OpenSpec.

## Impact

- Aligns assistant transcript storage and failure deduplication with the sanitized markdown rendering path.
- Prevents malformed model output from creating raw/storage drift or duplicate error noise.
