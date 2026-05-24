## Why

The cockpit dead-letter filter bar already supports latest family, subtype, and actor/time filters, but operators still cannot narrow by whether a family appeared at all without leaving the cockpit.

## What Changes

- add `any_match_family` to the cockpit filter bar
- forward the any-match family filter through the existing dead-letter list bridge
- document the landed any-match-family-filtering slice in public docs and main OpenSpec specs

## Impact

- better cockpit triage by observed retry lifecycle family
- no new backend endpoints
- no change to the current apply/reset model
