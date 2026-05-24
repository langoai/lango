## Why

The Dead Letters help bar already hides `Ctrl+R` while a retry request is actively running, but the in-page filter hint line still falls back to generic copy that advertises `Ctrl+R to reset`. That leaves two visible operator surfaces disagreeing about whether reset is actionable in the running state.

## What Changes

- Make the Dead Letters filter hint line omit `Ctrl+R` while a retry request is actively running.
- Add a regression for the running-state hint copy.
- Sync the docs/spec wording with the broader running-state contract.

## Impact

- Removes a visible truth drift from the Dead Letters retry flow.
- Keeps the help bar and inline hint copy aligned during retry execution.
