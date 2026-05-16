## Why

The cockpit Status page renders feature names, disable reasons, tool names, graph-admission sources, producer groups, and validator identifiers directly from runtime-fed data. Those strings are not normalized or stripped before display, so malformed values can leak control sequences or embedded newlines into a read-only operator page.

## What Changes

- Sanitize operator-facing Status page text fields before rendering them.
- Add regressions for malformed feature and graph-admission metadata.
- Record the plain-text Status page contract in OpenSpec and downstream docs.

## Impact

- Brings the cockpit Status page up to the same plain-text baseline as the hardened chat surfaces.
- Prevents malformed runtime metadata from destabilizing the operator dashboard.
