## Why

Mission Control already supports `d` to dismiss the selected proposed mission, and the public cockpit docs mention that key, but the in-product help bar does not expose it. That makes an actionable proposal-only control harder to discover than the docs suggest.

## What Changes

- Show a `d` dismiss binding in Mission Control help when the selected row is a proposed mission.
- Add regressions for proposal and non-proposal help states.
- Update the cockpit-pages spec to require the context-sensitive dismiss help exposure.

## Impact

- Operators can discover proposal dismissal directly from the Mission Control help bar.
- Runtime help, tests, docs, and spec align on the same proposal-action key surface.
