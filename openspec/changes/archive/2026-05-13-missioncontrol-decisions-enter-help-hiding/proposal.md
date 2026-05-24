## Why

Mission Control's help bar still advertises `Enter` while the decisions lane is focused, but the actual pending-approval key path does not use `Enter` as an approval action. In that state the help is overstating what the operator can do.

## What Changes

- Hide the generic `Enter` help while the decisions lane is focused and no proposal-accept path is active.
- Add regressions for decisions-focus vs composer/proposal help states.
- Update the cockpit-pages spec and feature docs to describe the reduced decisions-lane help surface.

## Impact

- The decisions lane stops advertising an inert `Enter` key.
- Runtime help, tests, docs, and spec align more closely with the actual approval interaction path.
