## Why

The Dead Letters page already hides `↑/↓` help when there are no backlog rows, but it still advertises row navigation when exactly one row exists. In that state the row-navigation keys are still inert.

## What Changes

- Show Dead Letters `↑/↓` help only when there are at least two backlog rows.
- Add regressions for the single-row help contract.
- Update cockpit-pages spec and feature docs to require the stricter threshold.

## Impact

- Single-row Dead Letters states stop advertising inert row-navigation keys.
- Runtime help, tests, docs, and spec align on the same row-navigation contract.
