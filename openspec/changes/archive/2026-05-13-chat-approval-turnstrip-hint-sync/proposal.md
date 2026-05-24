## Why

The approving-state help bar now correctly advertises `d/Esc` deny and `Ctrl+D` quit, but the turn-state strip hint still says only `choose a / s / d`. That leaves two visible chat surfaces describing different approval-state key contracts.

## What Changes

- Update the approving-state turn strip hint to match the current deny and quit affordances.
- Add a regression for the hint copy.
- Sync public docs/spec wording with the updated strip hint.

## Impact

- Removes a visible copy mismatch from the approval state.
- Keeps the turn strip aligned with the help bar and actual key handling.
