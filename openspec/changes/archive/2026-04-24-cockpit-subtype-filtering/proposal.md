## Why

The cockpit dead-letter filter bar is landed, but operators still cannot narrow the backlog by the latest retry/dead-letter phase without leaving the cockpit.

## What Changes

- add `latest_status_subtype` to the cockpit filter bar
- forward the subtype filter through the existing dead-letter list bridge
- document the landed subtype-filtering slice in public docs and main OpenSpec specs

## Impact

- better cockpit backlog triage
- no new backend endpoints
- no change to the existing apply/reset model
