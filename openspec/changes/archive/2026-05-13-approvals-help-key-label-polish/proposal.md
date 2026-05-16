## Why

The Approvals page now supports both `Tab` and `/` for section switching, but the rendered help key label still uses a compressed `tab,/` format. That is serviceable, but the in-product label can be cleaner without changing behavior.

## What Changes

- Change the Approvals help key label from `tab,/` to `tab /`.
- Update regressions and the approval-history-view spec to pin the cleaner label.
- Keep the public cockpit docs aligned with the same visible key wording.

## Impact

- The Approvals help bar reads more naturally while preserving the same keys.
- Runtime help, tests, docs, and spec all describe the same label contract.
