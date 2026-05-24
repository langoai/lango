## Why

Mission Control already documents that proposed mission rows can expose `Enter` to accept, `d` to dismiss, and `↑/↓` when another row exists. The current `ShortHelp()` implementation can replace `↑/k` with `enter: accept` when the selected proposed mission sits in a multi-row lane, so the visible help surface drops a still-actionable navigation key.

## What Changes

- Preserve the navigation bindings when a proposed mission row is selected and another mission row exists.
- Add a regression for the combined proposed-row help surface.
- Sync the Mission Control docs/spec wording with that preserved-navigation contract.

## Impact

- Removes a visible key-surface bug from Mission Control.
- Keeps runtime help aligned with the documented operator contract.
