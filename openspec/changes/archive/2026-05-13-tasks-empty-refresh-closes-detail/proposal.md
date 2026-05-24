## Why

The Tasks page can remain in detail mode after a refresh removes every task row. In that state the page renders the empty-task body, but the internal mode still advertises detail actions like `Enter` close detail even though there is no selected task left to close. That leaves an inert help surface and stale UI state behind.

## What Changes

- Close Tasks detail mode automatically when a refresh leaves the page with no selected task.
- Add a regression for the empty-after-refresh path.
- Update the public cockpit docs and task-surface spec to describe the reset behavior.

## Impact

- Removes a stale detail-mode state from the Tasks page.
- Keeps empty-state help and page mode aligned after background task lists change.
