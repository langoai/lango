## Why

Mission Control help has become more context-sensitive, but `Enter` still appears on ordinary mission rows even when it will not actually do anything. In current runtime behavior, `Enter` on the missions lane is actionable for proposal acceptance or specific workbench-empty starter flows, not for an already-accepted durable mission row.

## What Changes

- Hide `Enter` in Mission Control help whenever the current focus state has no actionable Enter behavior.
- Keep `Enter` for proposal acceptance, composer submission, and empty-workbench starter flows.
- Add regressions and update cockpit-pages/docs wording to match the stricter actionability contract.

## Impact

- Mission Control stops advertising an inert `Enter` key on ordinary mission rows.
- Help becomes a closer description of the currently actionable surface.
