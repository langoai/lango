## Why

The cockpit dead-letter filter bar already supports subtype, family, actor, and time narrowing, but operators still cannot triage by dead-letter reason text or dispatch reference without leaving the cockpit.

## What Changes

- add `dead_letter_reason_query` to the cockpit filter bar
- add `latest_dispatch_reference` to the cockpit filter bar
- forward both values through the existing dead-letter list bridge
- document the landed reason/dispatch-filtering slice in public docs and main OpenSpec specs

## Impact

- better cockpit triage for operator investigations
- no new backend endpoints
- no change to the current apply/reset model
