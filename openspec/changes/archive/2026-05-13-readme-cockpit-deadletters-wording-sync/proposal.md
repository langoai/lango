## Why

The runtime and public cockpit feature page now treat `Dead Letters` as an always-registered degraded page. But the README shortcut table still says the page exists only "when the dead-letter bridge is available".

That is now directly false for one of the most visible public entrypoints.

## What Changes

- Update the README cockpit shortcut table to describe Dead Letters as a degraded page that remains available while its bridge is unavailable.
- Extend downstream docs sync coverage so README follows the same current contract as the cockpit feature page.

## Impact

- README matches the current cockpit runtime and feature docs.
- Public cockpit messaging stays consistent across the repo's most visible entrypoints.
