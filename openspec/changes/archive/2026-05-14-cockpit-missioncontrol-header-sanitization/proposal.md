## Why

The cockpit Mission Control header renders active-agent, model/provider, context, metrics, and degraded-note summaries directly from runtime-fed projector data. Those values currently pass through without normalization, so malformed summaries can leak control sequences or embedded newlines into the default operator landing surface.

## What Changes

- Sanitize operator-facing Mission Control header summary text before rendering it.
- Add regression coverage for malformed header metadata.
- Record the plain-text Mission Control header contract in OpenSpec and downstream docs.

## Impact

- Brings the default Mission Control header up to the same plain-text baseline as the hardened chat and cockpit detail pages.
- Prevents malformed projector metadata from destabilizing the operator dashboard.
