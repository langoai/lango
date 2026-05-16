## Why

The Approvals page already hides inert navigation help in empty sections, but the grants section still advertises `r` and `R` whenever a grant store exists, even if there are no grant rows to act on. That leaves revoke actions discoverable when they cannot actually do anything.

## What Changes

- Hide `r` and `R` help in the Approvals grants section unless there is at least one grant row.
- Add regression coverage for empty and populated grants states.
- Update the approval-history-view spec and cockpit docs to describe the same action-help contract.

## Impact

- Empty grants states no longer advertise revoke actions that cannot run.
- Runtime help, tests, docs, and spec describe the same section-sensitive behavior.
