## Why

The Dead Letters page already renders explicit load-failure messages when the backlog query fails or when selected-transaction detail fails to load, and tests cover both behaviors. The main `cockpit-pages` spec currently pins only the unavailable-without-callback path, so it underspecifies a real operator-visible failure surface.

## What Changes

- Extend the main `cockpit-pages` spec with explicit backlog-load-failure and detail-load-failure scenarios for Dead Letters.

## Impact

- The main cockpit spec matches the actual Dead Letters failure surface more closely.
- Future regressions in load-failure wording become easier to catch at the spec layer.
