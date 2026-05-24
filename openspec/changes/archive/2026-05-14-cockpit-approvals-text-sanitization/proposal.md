## Why

The cockpit Approvals page renders approval history and active grant metadata directly from runtime-fed stores. Tool names, summaries, outcomes, providers, and session keys currently pass through without normalization, so malformed values can leak control sequences or embedded newlines into a read-only operator surface.

## What Changes

- Sanitize operator-facing Approvals page text before rendering it.
- Add regressions for malformed history and grant metadata.
- Record the plain-text Approvals page contract in OpenSpec and downstream docs.

## Impact

- Brings the cockpit Approvals page up to the same plain-text baseline as the hardened chat, Status, and Tools surfaces.
- Prevents malformed approval metadata from destabilizing the operator dashboard.
