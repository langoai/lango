## Why

The cockpit dead-letter filter bar already supports subtype, actor, and time filters, but operators still cannot narrow by the latest retry family without leaving the cockpit.

## What Changes

- add `latest_status_subtype_family` to the cockpit filter bar
- forward the latest-family filter through the existing dead-letter list bridge
- document the landed family-filtering slice in public docs and main OpenSpec specs

## Impact

- better cockpit triage by retry lifecycle family
- no new backend endpoints
- no change to the current apply/reset model
