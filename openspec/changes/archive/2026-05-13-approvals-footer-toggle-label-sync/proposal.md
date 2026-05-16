## Why

The Approvals page help bar already advertises both `Tab` and `/` for section switching as `tab /`, but the footer hint strip still shows only `[/] switch`. That leaves two visible help surfaces disagreeing about the accepted section-toggle keys.

## What Changes

- Update the Approvals footer hint strip to show `tab /` for section switching.
- Add a regression covering the footer label.
- Sync the approvals docs/spec wording with the footer/help agreement.

## Impact

- Removes another visible help-surface mismatch from the Approvals page.
- Keeps operator guidance aligned across the footer and help bar.
