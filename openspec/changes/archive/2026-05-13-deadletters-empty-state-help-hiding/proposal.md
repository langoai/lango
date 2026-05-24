## Why

The Dead Letters page help bar currently advertises `↑/↓` row navigation even when there is no backlog row to move through. In empty and unavailable states, that guidance is inert and less truthful than the rest of the cockpit help surfaces.

## What Changes

- Hide Dead Letters `↑/↓` row-navigation help when no backlog rows are present.
- Add regressions for empty and populated help states.
- Update cockpit-pages spec and feature docs to describe the reduced empty-state help surface.

## Impact

- Empty and unavailable Dead Letters states stop advertising inert row-navigation keys.
- Runtime help, tests, docs, and spec describe the same context-sensitive contract.
