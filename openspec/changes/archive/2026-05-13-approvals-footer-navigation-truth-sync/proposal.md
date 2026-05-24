## Why

The Approvals page already hides inert `↑/↓` bindings from `ShortHelp()` when the active section has zero or one row, but the page footer still always advertises `[↑/↓] navigate`. That leaves two visible help surfaces disagreeing about whether row navigation is actionable.

## What Changes

- Make the Approvals page footer advertise row navigation only when the active section actually has another row.
- Add a regression covering the single-row grants case.
- Sync the approvals spec/docs wording with the footer/help agreement.

## Impact

- Removes a visible help-surface mismatch from the Approvals page.
- Keeps footer hints aligned with actual keyboard actionability.
