## Why

Dead-letter operators can already inspect raw latest reason strings, actors, and dispatch references, but raw reason strings are noisy and require manual interpretation. A first grouped reason-family summary gives operators a faster high-level read without hiding the raw reason strings needed for diagnosis.

## What Changes

- Add grouped latest dead-letter reason-family summaries to `lango status dead-letter-summary`.
- Render the CLI grouped buckets as `by_reason_family` in JSON and as a `By reason family` table section.
- Add a cockpit `reason families:` line to the dead-letters page summary strip.
- Use the initial built-in taxonomy: `retry-exhausted`, `policy-blocked`, `receipt-invalid`, `background-failed`, and `unknown`.
- Preserve raw top latest dead-letter reasons alongside the grouped view.
- Update public docs and docs-only OpenSpec requirements.

## Impact

- Operators get a quicker backlog-level failure-family signal.
- Existing dead-letter summary fields remain additive and backward compatible.
- No Go control-plane behavior, retry semantics, backend read model, or filtering behavior changes are introduced by the docs archive.
