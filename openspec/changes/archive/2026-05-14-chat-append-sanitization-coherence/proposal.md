## Why

Several chat transcript paths already sanitize display text at append time, but `system`, `status`, and some display-only metadata paths still store raw strings and rely on render-time cleanup. That leaves the transcript model internally inconsistent and makes future render-path reuse easier to get wrong.

## What Changes

- Sanitize `system` and `status` entry content before storing it in the transcript model.
- Sanitize display-only channel/delegation metadata before storing it.
- Add regression coverage that appended transcript data is stored in display-safe form.
- Record the coherence contract in OpenSpec.

## Impact

- Aligns transcript entry storage with the already-hardened display contract.
- Reduces the chance of future regressions if alternate render paths reuse stored entry data.
