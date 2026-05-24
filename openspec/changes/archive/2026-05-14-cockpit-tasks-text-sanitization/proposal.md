## Why

The cockpit Tasks page renders prompt previews, statuses, origin labels, results, errors, and transient action messages directly from runtime-fed task data. Those values currently pass through without normalization, so malformed task metadata can leak control sequences or embedded newlines into a read-only operator surface.

## What Changes

- Sanitize operator-facing Tasks page text before rendering it.
- Sanitize transient task action status messages before display.
- Add regressions for malformed task table/detail/status-message content.
- Record the plain-text Tasks page contract in OpenSpec and downstream docs.

## Impact

- Brings the cockpit Tasks page up to the same plain-text baseline as the hardened chat, Status, Tools, and Approvals surfaces.
- Prevents malformed background-task metadata from destabilizing the operator dashboard.
