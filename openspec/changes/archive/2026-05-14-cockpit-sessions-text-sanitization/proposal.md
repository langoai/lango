## Why

The cockpit Sessions page renders session keys and configured-source load errors directly from runtime-fed data. Those values currently pass through without normalization, so malformed session metadata or backend error text can leak control sequences or embedded newlines into a read-only operator surface.

## What Changes

- Sanitize operator-facing Sessions page text before rendering it.
- Add regressions for malformed session keys and session-list load errors.
- Record the plain-text Sessions page contract in OpenSpec and downstream docs.

## Impact

- Brings the cockpit Sessions page up to the same plain-text baseline as the hardened chat and other cockpit surfaces.
- Prevents malformed session/runtime metadata from destabilizing the operator dashboard.
