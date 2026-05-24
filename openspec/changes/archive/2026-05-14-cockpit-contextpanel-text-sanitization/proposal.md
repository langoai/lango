## Why

The cockpit context panel renders top-tool names, runtime active-agent labels, and channel names directly from live runtime data. Those values currently pass through without normalization, so malformed labels can leak control sequences or embedded newlines into an always-visible operator surface.

## What Changes

- Sanitize operator-facing context-panel labels before rendering them.
- Add regression coverage for malformed tool, agent, and channel names.
- Record the plain-text context-panel contract in OpenSpec and downstream docs.

## Impact

- Brings the cockpit context panel up to the same plain-text baseline as the hardened page surfaces.
- Prevents malformed live runtime labels from destabilizing the persistent right-side metrics panel.
