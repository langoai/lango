## Why

Mission Control proposal rendering is hardened, but `LearningSuggestionBuffer` still stores raw `LearningSuggestionEvent` text fields. That leaves the fallback proposal producer and any future replay consumers dependent on downstream sanitization instead of the session-scoped learning buffer itself being display-safe.

## What Changes

- Sanitize display-facing learning suggestion fields when appending to the cockpit learning buffer.
- Add regression coverage for malformed buffered learning suggestion text.
- Record the learning-buffer replay-safety contract in OpenSpec and downstream docs.

## Impact

- Aligns the transient proposal producer path with the same plain-text baseline already enforced by projector/render stages.
- Prevents raw control text from persisting inside the cockpit’s learning suggestion buffer.
