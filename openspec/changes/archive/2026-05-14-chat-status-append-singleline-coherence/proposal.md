## Why

Compact status rows render as plain single-line text, but `appendStatus()` still stores stripped-yet-multiline content and relies on render-time normalization. That leaves the transcript model one step behind the compact status surface contract.

## What Changes

- Normalize appended status content to the same plain single-line form used by compact status rendering.
- Add regression coverage for stored multiline status content.
- Record the status-entry append-time coherence contract in OpenSpec.

## Impact

- Aligns stored status transcript data with the already-hardened compact row UI.
- Reduces the chance of future regressions if alternate code paths inspect or reuse stored status content.
