## Why

The cockpit Mission Control lanes still render mission titles/details, decision text, activity summaries, loop summaries, collaboration labels, and overflow summaries directly from runtime-fed projector data. Those values currently pass through without normalization, so malformed values can leak control sequences or embedded newlines into the default operator landing surface.

## What Changes

- Sanitize operator-facing Mission Control lane text before rendering it.
- Add regression coverage for malformed mission, decision, activity, and loop metadata.
- Record the plain-text Mission Control lane contract in OpenSpec and downstream docs.

## Impact

- Extends the Mission Control hardening beyond the header to the core dashboard lanes the operator sees first.
- Prevents malformed projector metadata from destabilizing the default cockpit landing surface.
