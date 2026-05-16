## Why

Mission Control's proposed-mission actions still have silent no-op edges:

- accepting a proposed mission does nothing when `missionService` is absent
- dismissing a proposed mission does nothing when `proposalSvc` is absent

Those are operator actions on visible UI affordances. Silent no-op is the wrong contract; the page should explain why the action cannot run.

## What Changes

- Make proposed-mission accept fail closed with an explicit system message when `missionService` is absent.
- Make proposed-mission dismiss fail closed with an explicit system message when `proposalSvc` is absent.
- Add regressions for both silent-no-op edges.
- Sync cockpit page spec and feature docs.

## Impact

- Mission Control no longer drops visible proposal actions on missing backing services.
- Operators get actionable explanations instead of nothing happening.
