## Why

The chat transcript now shows approval events with compact request-id annotations, but that traceability improvement is only described in the detailed cockpit feature docs. The README and CLI core overview still talk about approval surfaces more generically.

## What Changes

- Update the README to mention compact request-id annotations on approval transcript events.
- Update the CLI core overview to mention the same approval traceability improvement.
- Record the sync in OpenSpec.

## Impact

- Keeps first-touch docs aligned with the current approval transcript contract.
- Helps operators understand how repeated approvals stay distinguishable in chat history.
